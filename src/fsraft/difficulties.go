package fsraft

import (
	fs "filesystem"
	"testing"
)

// The combiner: takes a difficulty and a functionality test and runs that test on that difficulty.
func runFunctionalityTestWithDifficulty(t *testing.T, functionalityTest func(t *testing.T, f fs.FileSystem),
	difficulty func(t *testing.T) fs.FileSystem) {
	functionalityTest(t, difficulty(t))
}

// Difficulty settings ========================================================

// Whenever you add a new difficulty setting, be sure to add it to this list.
// This list is used in to run every functionality test on every difficulty.
var Difficulties = []func(t *testing.T) fs.FileSystem{
	OneClerkThreeServersNoErrors,
   OneClerkFiveServersErrors,
}

func OneClerkThreeServersNoErrors(t *testing.T) fs.FileSystem {
	cfg := make_config(t, 3, false, -1)
	return cfg.makeClient(cfg.All())
}

func OneClerkFiveServersErrors(t *testing.T) fs.FileSystem {
	cfg := make_config(t, 5, true, -1)
	return cfg.makeClient(cfg.All())
}

/*
func OneClerkFiveServersNetworkPartition(t *testing.T) fs.FileSystem {
   ...
}
*/
