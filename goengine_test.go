package goengine

import (
	"reflect"
	"sort"
	"testing"
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
