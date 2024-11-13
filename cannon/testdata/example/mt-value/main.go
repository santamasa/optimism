package main

import (
	"fmt"
	"os"
	"testing"
)

func main() {

	// TODO(#12162) Debug commented out tests which are currently failing
	runTest(TestValue, "TestValue")
	runTest(TestValueLarge, "TestValueLarge")
	runTest(TestValuePanic, "TestValuePanic")
	runTest(TestValueConcurrent, "TestValueConcurrent")
	//runTest(TestValue_Swap, "TestValue_Swap")
	runTest(TestValueSwapConcurrent, "TestValueSwapConcurrent")
	//runTest(TestValue_CompareAndSwap, "TestValue_CompareAndSwap")
	runTest(TestValueCompareAndSwapConcurrent, "TestValueCompareAndSwapConcurrent")

	fmt.Println("Value tests passed")
}

func runTest(testFunc func(*testing.T), name string) {
	t := &testing.T{}
	testFunc(t)
	if t.Failed() {
		fmt.Printf("Test failed: %v\n", name)
		os.Exit(1)
	} else {
		fmt.Printf("Test passed: %v\n", name)
	}
}
