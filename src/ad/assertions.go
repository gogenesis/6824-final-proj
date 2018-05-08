package ad

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func Assert(cond bool) {
	// Skip assertions when debugging is turned off for performance
	if assertionsEnabled && (!cond) {
		panic("Assertion error!")
	}
}

func AssertEquals(expected, actual interface{}) {
	// Skip assertions when debugging is turned off for performance
	if assertionsEnabled && (expected != actual) {
		panic(fmt.Sprintf("AssertionError: expected %v, got %v\n", expected, actual))
	}
}

func AssertT(t *testing.T, cond bool) {
	if !cond {
		failWithMessageAndStackTrace(t, "Assertion error!")
	}
}

func AssertNoErrorT(t *testing.T, e error) {
	if e != nil {
		failWithMessageAndStackTrace(t, "Assertion error! Expected no error, got \"%v\"", e.Error())
	}
}

func AssertEqualsT(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		failWithMessageAndStackTrace(t, "Assertion error! Expected %+v, got %+v\n", expected, actual)
	}
}

func AssertValidFDT(t *testing.T, fd int) {
	if fd < 3 {
		failWithMessageAndStackTrace(t, fmt.Sprintf("Impossible file descriptor %d.", fd))
	}
}

// Given an object, asserts that it is an error and returns it, or asserts that it is nil and returns it.
// If it is neither an error nor nil, panics.
func AssertIsErrorOrNil(object interface{}) error {
	if object == nil {
		return nil
	}
	return object.(error)
}

func AssertExplainT(t *testing.T, cond bool, format string, a ...interface{}) {
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
