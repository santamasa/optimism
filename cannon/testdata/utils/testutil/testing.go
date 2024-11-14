package testutil

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
)

func RunTest(testFunc func(testing.TB), name string) {
	runner := newTestRunner(name)
	runner.Run("", testFunc)
}

func ExecRunnerTest(testFunc func(*TestRunner), name string) {
	runner := newTestRunner(name)
	testFunc(runner)
}

type TestRunner struct {
	*mockT
	baseName string
}

func newTestRunner(baseName string) *TestRunner {
	return &TestRunner{mockT: newMockT(), baseName: baseName}
}

func (r *TestRunner) Run(name string, f func(t testing.TB)) bool {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			recover() // Recover in case of runtime.Goexit()

			testName := r.baseName
			if name != "" {
				testName = fmt.Sprintf("%v (%v)", r.baseName, name)
			}

			if r.Failed() {
				fmt.Printf("Test failed: %v\n", testName)
				os.Exit(1)
			} else if r.Skipped() {
				fmt.Printf("Test skipped: %v\n", testName)
			} else {
				fmt.Printf("Test passed: %v\n", testName)
			}

			wg.Done()
		}()

		f(r)
	}()

	wg.Wait()

	return !r.failed
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
