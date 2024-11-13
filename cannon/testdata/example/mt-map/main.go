package main

import (
	"fmt"
	"os"
	"testing"
)

func main() {
	runTest(TestMapMatchesRWMutex, "TestMapMatchesRWMutex")
	runTest(TestMapMatchesDeepCopy, "TestMapMatchesDeepCopy")
	runTest(TestConcurrentRange, "TestConcurrentRange")
	runTest(TestIssue40999, "TestIssue40999")
	runTest(TestMapRangeNestedCall, "TestMapRangeNestedCall")
	runTest(TestCompareAndSwap_NonExistingKey, "TestCompareAndSwap_NonExistingKey")
	runTest(TestMapRangeNoAllocations, "TestMapRangeNoAllocations")

	fmt.Println("Map test passed")
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
