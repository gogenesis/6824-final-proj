package fsraft

import (
	"testing"
)

// The combiner: takes a "difficulty" and a "test" and runs that test on that difficulty.
func runUnitTestWithDifficulty(t *testing.T, difficulty func(t *testing.T) FileSystem, testToRun func(fs FileSystem)) {
	fs := difficulty(t)
	testToRun(fs)
}

// Difficulty settings ========================================================

func OneClerkOneServerNoErrors(t *testing.T) FileSystem {
	cfg := make_config(t, 1, false, -1)
	return cfg.makeClient(cfg.All())
}

func OneClerkFiveServersNoErrors(t *testing.T) FileSystem {
	cfg := make_config(t, 5, false, -1)
	return cfg.makeClient(cfg.All())
}

// Combinations of unit tests and difficulty settings =========================
// Unit tests themselves are in filesystem_test.go.
// I (David) am looking into how to autogenerate these for every combination of unit test and diffculty,
// but in the meantime, we can write them by hand.

func TestClerk_OneClerkOneServerNoErrors_OpenClose(t *testing.T) {
	runUnitTestWithDifficulty(t, OneClerkOneServerNoErrors, TestFSOpenClose)
}

func TestClerk_OneClerkFiveServersNoErrors_OpenClose(t *testing.T) {
	runUnitTestWithDifficulty(t, OneClerkFiveServersNoErrors, TestFSOpenClose)
}

func TestClerk_OneClerkOneServerNoErrors_BasicReadWrite(t *testing.T) {
	runUnitTestWithDifficulty(t, OneClerkOneServerNoErrors, TestFSBasicReadWrite)
}

func TestClerk_OneClerkFiveServersNoErrors_BasicReadWrite(t *testing.T) {
	runUnitTestWithDifficulty(t, OneClerkFiveServersNoErrors, TestFSBasicReadWrite)
}


// TODO when more tests are written in filesystem_tests.go, call them here.
// I (David) will look into
