package main

import (
	"fmt"
	"testing"

	"utils/testutil"
)

func main() {
	testutil.ExecRunnableTest(ShouldFail, "ShouldFail")

	fmt.Println("Passed test that should have failed")
}

func ShouldFail(t *testutil.TestRunner) {
	t.Run("", func(t testing.TB) {
		t.Fail()
	})
}
