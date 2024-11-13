package main

import (
	"fmt"
	"os"
	"testing"
)

func main() {
	runTest(TestOnceFunc, "TestOnceFunc")
	runTest(TestOnceValue, "TestOnceValue")
	runTest(TestOnceValues, "TestOnceValues")
	runTest(TestOnceFuncPanic, "TestOnceFuncPanic")
	runTest(TestOnceValuePanic, "TestOnceValuePanic")
	runTest(TestOnceValuesPanic, "TestOnceValuesPanic")
	runTest(TestOnceFuncPanicNil, "TestOnceFuncPanicNil")
	runTest(TestOnceFuncGoexit, "TestOnceFuncGoexit")

	// TODO(#12162) Debug the following tests
	//runTest(TestOnceFuncPanicTraceback, "TestOnceFuncPanicTraceback")
	//runTest(TestOnceXGC, "TestOnceXGC")

	fmt.Println("OnceFunc tests passed")
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
