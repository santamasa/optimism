package main

import (
	"fmt"
	"os"
	"testing"
)

func main() {
	runTest(TestWaitGroup, "TestWaitGroup")
	runTest(TestWaitGroupMisuse, "TestWaitGroupMisuse")
	runTest(TestWaitGroupRace, "TestWaitGroupRace")
	runTest(TestWaitGroupAlign, "TestWaitGroupAlign")

	fmt.Println("WaitGroup tests passed")
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
