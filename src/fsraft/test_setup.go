package fsraft

import (
	"fmt"
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

// You can change these from panic to t.Fatalf if it would make your life easier
func assertNoErrorPanic(e error) {
	if e != nil {
		panic(e.Error())
	}
}

func assertEqualsPanic(expected, actual interface{}) {
	if expected != actual {
		panic(fmt.Sprintf("Assertion error! Expected %+v, got %+v\n", expected, actual))
	}
}

func assertPanic(cond bool) {
	if !cond {
		panic("Assertion error!")
	}
}

func assertFail(t *testing.T, cond bool) {
	if !cond {
		t.Fatalf("Assertion error!")
	}
}

func assertNoErrorFail(t *testing.T, e error) {
	if e != nil {
		t.Fatalf(e.Error())
	}
}

func assertEqualsFail(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		t.Fatalf("Assertion error! Expected %+v, got %+v\n", expected, actual)
	}
}

// Code generation ============================================================

// Someone on the
