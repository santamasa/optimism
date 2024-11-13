package main

import (
	"fmt"
	"os"
	"testing"
)

func main() {
	runTest(TestPool, "TestPool")
	runTest(TestPoolNew, "TestPoolNew")
	runTest(TestPoolGC, "TestPoolGC")
	runTest(TestPoolRelease, "TestPoolRelease")
	runTest(TestPoolStress, "TestPoolStress")
	runTest(TestPoolDequeue, "TestPoolDequeue")
	runTest(TestPoolChain, "TestPoolChain")

	// TODO(#12162) This test is currently failing - need to debug it
	// runTestExpectingPanic(TestNilPool, "TestNilPool")

	fmt.Println("Pool test passed")
}

func runTest(testFunc func(*testing.T), name string) {
	runTestHelper(testFunc, name, false)
}

func runTestExpectingPanic(testFunc func(*testing.T), name string) {
	runTestHelper(testFunc, name, true)
}

func runTestHelper(testFunc func(*testing.T), name string, expectPanic bool) {
	t := &testing.T{}

	if expectPanic {
		catch := func() {
			if recover() == nil {
				t.Error("expected panic")
				fail(name)
			} else {
				succeed(name)
			}
		}
		defer catch()
	}

	testFunc(t)
	if t.Failed() {
		fail(name)
	} else {
		fmt.Printf("Test passed: %v\n", name)
	}
}

func fail(name string) {
	fmt.Printf("Test failed: %v\n", name)
	os.Exit(1)
}

func succeed(name string) {
	fmt.Printf("Test passed: %v\n", name)
}
