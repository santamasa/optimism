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
	testutil.ExecRunnableTest(TestValue_Swap, "TestValue_Swap")
	testutil.RunTest(TestValueSwapConcurrent, "TestValueSwapConcurrent")
	testutil.ExecRunnableTest(TestValue_CompareAndSwap, "TestValue_CompareAndSwap")
	testutil.RunTest(TestValueCompareAndSwapConcurrent, "TestValueCompareAndSwapConcurrent")

	fmt.Println("Value tests passed")
}
