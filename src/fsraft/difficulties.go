package fsraft

import (
   "time"
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

// ===== proposed network partition test ======
const electionTimeout = 1 * time.Second

// UNVERIFIED - requesting review
// Omitted the word Test fully so it doesn't get picked up yet
func TesTmpTOneClerkFiveServersNetworkPartition(t *testing.T) {
   const nservers = 5
	cfg := make_config(t, nservers, false, -1)
	defer cfg.cleanup()
	ck := cfg.makeClient(cfg.All())

   // Perhaps Put => fs.Write ...
	// Equivalent to Put(cfg, ck, "1", "13")
   fd := fs.HelpOpen(t, ck, "/1", fs.ReadWrite, fs.Create)
   fs.HelpWriteBytes(t, ck, fd, []byte("13")) //offset now 2
   fs.HelpClose(t, ck, fd)

	cfg.begin("Test: progress in majority")

	p1, p2 := cfg.make_partition()
	cfg.partition(p1, p2)

	ckp1 := cfg.makeClient(p1)  // connect ckp1 to p1
	ckp2a := cfg.makeClient(p2) // connect ckp2a to p1
	ckp2b := cfg.makeClient(p2) // connect ckp2b to p2

	//Equivalent to Put(cfg, ckp1, "1", "14")
   fd1 := fs.HelpOpen(t, ckp1, "/1", fs.ReadWrite, fs.Create)
   fs.HelpSeek(t, ckp1, fd1, 0, fs.FromBeginning)
   fs.HelpWriteBytes(t, ckp1, fd1, []byte("14")) //offset now 2
   fs.HelpClose(t, ckp1, fd1)

   // Perhaps check / Get => fs.Read ...
   // Equivalent to check(cfg, t, ckp1, "1", "14")
   fd1 = fs.HelpOpen(t, ckp1, "/1", fs.ReadOnly, fs.Create)
   fs.HelpSeek(t, ckp1, fd1, 0, fs.FromBeginning)
   _, dataCkp1 := fs.HelpRead(t, ckp1, fd1, 2)
   fs.HelpVerifyBytes(t, dataCkp1, []byte("14"), "ckp1 read one")

	cfg.end()

	done0 := make(chan bool)
	done1 := make(chan bool)

	cfg.begin("no progress in minority (network partition)")
	go func() {
		// Equivalent to Put(cfg, ckp2a, "1", "15")
      fd2 := fs.HelpOpen(t, ckp2a, "/1", fs.ReadWrite, fs.Create)
      fs.HelpSeek(t, ckp2a, fd2, 0, fs.FromBeginning)
      fs.HelpWriteBytes(t, ckp2a, fd2, []byte("15")) //offset now 2
      fs.HelpClose(t, ckp2a, fd2)
		done0 <- true
	}()
	go func() {
      // Equivalent to Get(cfg, ckp2b, "1") // different clerk in p2
      // Perhaps Get => fs.Seek, fs.Read
      fd2 := fs.HelpOpen(t, ckp2b, "/1", fs.ReadOnly, fs.Create)
      fs.HelpSeek(t, ckp2b, fd2, 0, fs.FromBeginning)
      _, _ = fs.HelpRead(t, ckp2b, fd2, 2)
      // not sure what the data should be here yet, original test
      // just does the Get and doesn't look at data
      //fs.HelpVerifyBytes(t, dataCkp2b, []byte("??"), "ckp1 read 1.5")
      fs.HelpClose(t, ckp2b, fd2)
		done1 <- true
	}()

	select {
	case <-done0:
		t.Fatalf("Write in minority completed")
	case <-done1:
		t.Fatalf("Read in minority completed")
	case <-time.After(time.Second):
	}

	//Equivalent to check(cfg, t, ckp1, "1", "14")
   fd = fs.HelpOpen(t, ckp1, "/1", fs.ReadOnly, fs.Create)
   fs.HelpSeek(t, ckp1, fd, 0, fs.FromBeginning)
   _, dataCkp1 = fs.HelpRead(t, ckp1, fd, 2) //offset now 2
   fs.HelpVerifyBytes(t, dataCkp1, []byte("14"), "ckp1 read two")
   fs.HelpClose(t, ckp1, fd)

	// Eqivalent to Put(cfg, ckp1, "1", "16")
   fd = fs.HelpOpen(t, ckp1, "/1", fs.ReadWrite, fs.Create)
   fs.HelpSeek(t, ckp1, fd, 0, fs.FromBeginning)
   fs.HelpWriteBytes(t, ckp1, fd, []byte("16")) //offset now 2
   fs.HelpClose(t, ckp1, fd)

	//Equivalent to check(cfg, t, ckp1, "1", "16")
   fd = fs.HelpOpen(t, ckp1, "/1", fs.ReadOnly, fs.Create)
   fs.HelpSeek(t, ckp1, fd, 0, fs.FromBeginning)
   _, dataCkp1 = fs.HelpRead(t, ckp1, fd, 2) //offset now 2
   fs.HelpVerifyBytes(t, dataCkp1, []byte("16"), "ckp1 read three")
   fs.HelpClose(t, ckp1, fd)

   // This followed, not sure what this checks in original test
	//check(cfg, t, ckp1, "1", "14") //check MemoryFS state

	cfg.end()

	cfg.begin("Test: completion after heal")

	cfg.ConnectAll()
	cfg.ConnectClient(ckp2a, cfg.All())
	cfg.ConnectClient(ckp2b, cfg.All())

	time.Sleep(electionTimeout)

	select {
	case <-done0:
	case <-time.After(30 * 100 * time.Millisecond):
		t.Fatalf("Write did not complete")
	}

	select {
	case <-done1:
	case <-time.After(30 * 100 * time.Millisecond):
		t.Fatalf("Read did not complete")
	default:
	}

	//Equivalent to check(cfg, t, ck, "1", "15")
   fd = fs.HelpOpen(t, ck, "/1", fs.ReadOnly, fs.Create)
   fs.HelpSeek(t, ck, fd, 0, fs.FromBeginning)
   _, dataCk := fs.HelpRead(t, ck, fd, 2) //offset now 2
   fs.HelpVerifyBytes(t, dataCk, []byte("15"), "ckp1 read three")
   fs.HelpClose(t, ck, fd)

	cfg.end()
   return
}
