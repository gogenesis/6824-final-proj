package fsraft

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// Ihis file facilitates unit tests, but does not run any tests itself.
// It includes difficulty settings for distributed tests and assert statements.
// This program runs generate_unit_tests.go when you type "go generate" because of the following magic comment:
//go:generate go run generate_unit_tests.go

// The combiner: takes a difficulty and a functionality test and runs that test on that difficulty.
func runFunctionalityTestWithDifficulty(t *testing.T, functionalityTest func(t *testing.T, fs FileSystem), difficulty func(t *testing.T) FileSystem) {
	functionalityTest(t, difficulty(t))
}

// Difficulty settings ========================================================

// Whenever you add a new difficulty setting, be sure to add it to this list.
// This list is used in to run every functionality test on every difficulty.
var Difficulties = []func(t *testing.T) FileSystem{
	OneClerkOneServerNoErrors,
	OneClerkFiveServersNoErrors,
}

func OneClerkOneServerNoErrors(t *testing.T) FileSystem {
	cfg := make_config(t, 1, false, -1)
	return cfg.makeClient(cfg.All())
}

func OneClerkFiveServersNoErrors(t *testing.T) FileSystem {
	cfg := make_config(t, 5, false, -1)
	return cfg.makeClient(cfg.All())
}

// Assertions =================================================================

func assert(t *testing.T, cond bool) {
	if !cond {
		failWithMessageAndStackTrace(t, "Assertion error!")
	}
}

func assertNoError(t *testing.T, e error) {
	if e != nil {
		failWithMessageAndStackTrace(t, "Assertion error! Expected no error, got \"%v\"", e.Error())
	}
}

func assertEquals(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		failWithMessageAndStackTrace(t, "Assertion error! Expected %+v, got %+v\n", expected, actual)
	}
}

func assertValidFD(t *testing.T, fd int) {
	if fd < 3 {
		failWithMessageAndStackTrace(t, fmt.Sprintf("Impossible file descriptor %d.", fd))
	}
}

func assertExplain(t *testing.T, cond bool, format string, a ...interface{}) {
	if !cond {
		failWithMessageAndStackTrace(t, "Assertion error! %s", fmt.Sprintf(format, a...))
	}
}

func failWithMessageAndStackTrace(t *testing.T, format string, a ...interface{}) {
	// skip StackTrace, this function, and the assert function
	t.Fatalf("%v\n%v", fmt.Sprintf(format, a...), stackTrace(3))
}

// Prints the stack trace, not including this function, to stdout.
func printStackTrace() {
	// skip=2 to skip two functions: StackTrace and printStackTrace.
	fmt.Println(stackTrace(2))

}

// Get a string of the stack trace, skipping skip calls at the bottom.
// For example, in StackTrace(0), the top function call in the stack is the call to StackTrace.
// The skip parameter exists because you might want to start reading somewhere more interesting.
func stackTrace(skip int) string {
	buf := make([]byte, 1<<16)                // idk, this is probably enough
	bytesWritten := runtime.Stack(buf, false) // false for only this goroutine
	buf = buf[:bytesWritten]                  // remove trailing 0 bits
	lines := strings.Split(string(buf), "\n")
	// Remove one line of header and two lines per function to skip
	lines = lines[1+2*skip:]
	return strings.Join(lines, "\n")

}

// Code generation ============================================================

// Someone on the
