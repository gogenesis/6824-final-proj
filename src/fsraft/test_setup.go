package fsraft

import (
	"fmt"
	"testing"
)

// Ihis file facilitates unit tests, but does not run any tests itself.
// It includes difficulty settings for distributed tests and assert statements.
// This program runs generate_combination_tests.go when you type "go generate" because of the following magic comment:
//go:generate go run generate_combination_tests.go

// The combiner: takes a difficulty and a functionality test and runs that test on that difficulty.
func runFunctionalityTestWithDifficulty(t *testing.T, functionalityTest func(fs FileSystem), difficulty func(t *testing.T) FileSystem) {
	functionalityTest(difficulty(t))
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
func assertNoError(e error) {
	if e != nil {
		panic(e.Error())
	}
}

func assertEquals(expected, actual interface{}) {
	if expected != actual {
		panic(fmt.Sprintf("Assertion error! Expected %+v, got %+v\n", expected, actual))
	}
}

func assert(cond bool) {
	if !cond {
		panic("Assertion error!")
	}
}

// Code generation ============================================================

// Someone on the
