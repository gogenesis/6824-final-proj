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
	OneClerkFiveServersUnreliableNet,
	OneClerkThreeServersSnapshots,
}

func OneClerkThreeServersNoErrors(t *testing.T) fs.FileSystem {
	cfg := make_config(t, 3, false, -1)
	return cfg.makeClient(cfg.All())
}

func OneClerkFiveServersUnreliableNet(t *testing.T) fs.FileSystem {
	cfg := make_config(t, 5, true, -1)
	return cfg.makeClient(cfg.All())
}

func OneClerkThreeServersSnapshots(t *testing.T) fs.FileSystem {
	cfg := make_config(t, 3, true, 1000) // arbitrarily
	return cfg.makeClient(cfg.All())
}

func OneClerkFiveServersRandomPartitions(t *testing.T) fs.FileSystem {
	// TODO return a wrapper class around a Clerk that delays ops and adds random partitions
	panic("TODO")
}

func TwoClerksThreeServersNoErrors(t *testing.T) fs.FileSystem {
	// TODO return a wrapper around a Clerk that (randomly?) distributes ops across two clerks
	panic("TODO")
}
