package main

import (
	"fmt"

	"utils/testutil"
)

func main() {
	testutil.RunTest(TestOnceFunc, "TestOnceFunc")
	testutil.RunTest(TestOnceValue, "TestOnceValue")
	testutil.RunTest(TestOnceValues, "TestOnceValues")
	testutil.RunTest(TestOnceFuncPanic, "TestOnceFuncPanic")
	testutil.RunTest(TestOnceValuePanic, "TestOnceValuePanic")
	testutil.RunTest(TestOnceValuesPanic, "TestOnceValuesPanic")
	testutil.RunTest(TestOnceFuncPanicNil, "TestOnceFuncPanicNil")
	testutil.RunTest(TestOnceFuncGoexit, "TestOnceFuncGoexit")

	// TODO(#12162) Debug the following tests
	//runTest(TestOnceFuncPanicTraceback, "TestOnceFuncPanicTraceback")
	//runTest(TestOnceXGC, "TestOnceXGC")

	fmt.Println("OnceFunc tests passed")
}
