package goengine

import (
	//"log"
	"reflect"
	"sort"
	"testing"
	"time"
)

func storeNum(eng *Engine, sk StateKey, num interface{}) Response {
	nums := eng.Get(sk).([]int)
	nums = append(nums, num.(int))
	eng.Set(sk, nums)
	return Response{Data: nil, Err: nil}
}

// TestStoreNums() demonstrates how to use engine.ActAndCloseRes() to dispatch
// a function that modifies the state and whose response is irrelevant.
func TestStoreNums(t *testing.T) {
	vals := []int{0, 1, 2, 3, 4}

	// Create and run the engine
	engine := BuildEngine()
	engine.Run()

	// Register the storeNum reducer function and an initial state
	numsSk, numsRk := engine.Register(make([]int, 0), storeNum)

	// For each number, make the engine dispatch storeNum() to store the number
	// in the state.
	for _, val := range vals {
		engine.Act(numsRk, val)
	}

	// Use the given StateKey to fetch data from the engine state
	nums := engine.Get(numsSk).([]int)

	// Sort nums as they may have been dispatched out of order
	sort.Ints(nums)

	if !reflect.DeepEqual(nums, vals) {
		t.Error("Numbers stored are not the same as those retrieved from the state")
	}
}

func incrementNum(eng *Engine, sk StateKey, num interface{}) Response {
	return Response{Data: num.(int) + 1, Err: nil}
}

// TestIncrementNums() demonstrates how to use engine.Act() to dispatch
// a function that does not modify the state but returns a result via a
// ResponseChan.
func TestIncrementNums(t *testing.T) {
	vals := []int{0, 1, 2, 3, 4}
	expected := []int{1, 2, 3, 4, 5}

	// Create and run the engine
	engine := BuildEngine()
	engine.Run()
	_, rk := engine.Register(nil, incrementNum)

	result := make([]int, 0)
	for _, val := range vals {
		response, _ := engine.Act(rk, val)
		result = append(result, response.Data.(int))
	}

	sort.Ints(result)

	if !reflect.DeepEqual(result, expected) {
		t.Error("Something went wrong when incrementing nums")
	}
}

func collectNums(eng *Engine, sk StateKey, numsSk interface{}) Response {
	//log.Println("Collecting...")
	collected := make([]int, 0)
	for i := 0; i < 100; i++ {
		val := eng.Get(numsSk.(StateKey)).(int)
		collected = append(collected, val)
		//log.Println("Collected:", val)
		sleepDuration, _ := time.ParseDuration("8ms")
		time.Sleep(sleepDuration)
	}
	//log.Println("Collected")
	return Response{Data: collected, Err: nil}
}

func modifyNums(eng *Engine, sk StateKey, data interface{}) Response {
	go func() {
		//log.Println("Modifying...")
		for i := 0; i < 100; i++ {
			eng.Mux.Lock()

			eng.UnsafeSet(sk, i)
			//log.Println("Modified:", i)
			sleepDuration, _ := time.ParseDuration("10ms")
			time.Sleep(sleepDuration)
			//log.Println("Reset")
			eng.UnsafeSet(sk, 0)

			eng.Mux.Unlock()
		}
		//log.Println("Modified")
	}()
	return Response{nil, nil}
}

func TestAtomicity(t *testing.T) {
	engine := BuildEngine()
	engine.Run()
	modifySk, modifyRk := engine.Register(0, modifyNums)
	_, collectRk := engine.Register(0, collectNums)

	engine.Act(modifyRk, nil)
	collectResponse, _ := engine.Act(collectRk, modifySk)
	collected := collectResponse.Data.([]int)

	if len(collected) != 100 {
		t.Error("Concurrent read/write failed")
	}

	for _, val := range collected {
		if val != 0 {
			t.Error("Concurrent read/write failed")
		}
	}
	//for {
	//}
}
