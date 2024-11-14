package testutil

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
)

func RunTest(testFunc func(testing.TB), name string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		t := newMockT()
		defer func() {
			recover() // Recover in case of runtime.Goexit()

			if t.Failed() {
				fmt.Printf("Test failed: %v\n", name)
				os.Exit(1)
			} else if t.Skipped() {
				fmt.Printf("Test skipped: %v\n", name)
			} else {
				fmt.Printf("Test passed: %v\n", name)
			}

			wg.Done()
		}()

		testFunc(t)
	}()

	wg.Wait()
}

type mockT struct {
	*testing.T
	failed  bool
	skipped bool
}

var _ testing.TB = (*mockT)(nil)

func newMockT() *mockT {
	return &mockT{}
}

func (t *mockT) Cleanup(func()) {}

func (t *mockT) Error(args ...any) {
	fmt.Print(args...)
	t.fail()
}

func (t *mockT) Errorf(format string, args ...any) {
	fmt.Printf(format, args...)
	t.fail()
}

func (t *mockT) Fail() {
	t.fail()
}

func (t *mockT) FailNow() {
	fmt.Println("Fatal")
	t.fail()
}

func (t *mockT) Failed() bool {
	return t.failed
}

func (t *mockT) Fatal(args ...any) {
	fmt.Print(args...)
	t.fail()
}

func (t *mockT) Fatalf(format string, args ...any) {
	fmt.Printf(format, args...)
	t.fail()
}

func (t *mockT) Helper() {}

func (t *mockT) Log(args ...any) {
	fmt.Print(args...)
}

func (t *mockT) Logf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (t *mockT) Name() string {
	return ""
}

func (t *mockT) Setenv(key, value string) {}

func (t *mockT) Skip(args ...any) {
	fmt.Println(args...)
	t.skip()
}

func (t *mockT) SkipNow() {
	t.skip()
}

func (t *mockT) Skipf(format string, args ...any) {
	fmt.Printf(format, args...)
	t.skip()
}
func (t *mockT) Skipped() bool {
	return t.skipped
}

func (t *mockT) skip() {
	t.skipped = true
	runtime.Goexit()
}

func (t *mockT) fail() {
	t.failed = true
	runtime.Goexit()
}

func (t *mockT) TempDir() string {
	t.Fatalf("TempDir not supported")
	return ""
}
