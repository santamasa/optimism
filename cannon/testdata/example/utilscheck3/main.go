package main

import (
	"fmt"

	"utils/testutil"
)

func main() {
	testutil.ExecRunnerTest(ShouldFail, "ShouldFail")

	fmt.Println("Passed test that should have failed")
}

func ShouldFail(t *testutil.TestRunner) {
	t.Fail()
}
