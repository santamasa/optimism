package main

import (
	"fmt"

	"utils/testutil"
)

func main() {
	testutil.RunTest(TestValue, "TestValue")
	testutil.RunTest(TestValueLarge, "TestValueLarge")
	testutil.RunTest(TestValuePanic, "TestValuePanic")
	testutil.RunTest(TestValueConcurrent, "TestValueConcurrent")
	testutil.ExecRunnerTest(TestValue_Swap, "TestValue_Swap")
	testutil.RunTest(TestValueSwapConcurrent, "TestValueSwapConcurrent")
	testutil.ExecRunnerTest(TestValue_CompareAndSwap, "TestValue_CompareAndSwap")
	testutil.RunTest(TestValueCompareAndSwapConcurrent, "TestValueCompareAndSwapConcurrent")

	fmt.Println("Value tests passed")
}
