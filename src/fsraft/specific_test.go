package fsraft

import (
	fs "filesystem"
	"fmt"
	"sync"
	"testing"
	"time"
)

// ===== proposed network partition test ======
const electionTimeout = 1 * time.Second

// UNVERIFIED - requesting review
// Omitted the word Test fully so it doesn't get picked up yet
//func TesTmpTOneClerkFiveServersNetworkPartition(t *testing.T) {
func TestOneClerkFiveServersPartition(t *testing.T) {
	const nservers = 5
	cfg := make_config(t, nservers, false, -1)
	defer cfg.cleanup()
	clerk := cfg.makeClient(cfg.All())
	dataFile := "/myFile.txt"

	// Equivalent to Put(cfg, clerk, "1", "13")
	fd := fs.HelpOpen(t, clerk, dataFile, fs.ReadWrite, fs.Create)
	fs.HelpWriteBytes(t, clerk, fd, []byte("13")) //offset now 2
	fs.HelpClose(t, clerk, fd)

	cfg.begin("Test: progress in majority")

	majority, minority := cfg.make_partition()
	cfg.partition(majority, minority)

	majorityClerk := cfg.makeClient(majority)
	minorityClerkA := cfg.makeClient(minority)
	minorityClerkB := cfg.makeClient(minority)

	//Equivalent to Put(cfg, clerkMajority, "1", "14")
	fd = fs.HelpOpen(t, majorityClerk, dataFile, fs.ReadWrite, 0)
	fs.HelpSeek(t, majorityClerk, fd, 0, fs.FromBeginning)
	fs.HelpWriteBytes(t, majorityClerk, fd, []byte("14")) //offset now 2
	fs.HelpClose(t, majorityClerk, fd)

	// Perhaps check / Get => fs.Read ...
	// Equivalent to check(cfg, t, clerkMajority, "1", "14")
	fd = fs.HelpOpen(t, majorityClerk, dataFile, fs.ReadOnly, 0)
	fs.HelpSeek(t, majorityClerk, fd, 0, fs.FromBeginning)
	_, data := fs.HelpRead(t, majorityClerk, fd, 2)
	fs.HelpVerifyBytes(t, data, []byte("14"),
		fmt.Sprintf("Error: Wrote 14 in majority, expected to read 14, instead read %v", data))
	fs.HelpClose(t, majorityClerk, fd)

	cfg.end()

	doneWithWriteInMinority := make(chan bool)
	doneWithReadInMinority := make(chan bool)

	cfg.begin("no progress in minority (network partition)")
	// These should complete after the partition is healed, but not before.
	// This lock ensures that each of these sets of operations is atomic.
	// This wasn't needed in lab 3 because there was only one operation done there,
	// so it was automatically atomic. Here, if both try to open the file at the
	// same time, one will fail with AlreadyOpen.
	var minorityLock sync.Mutex
	go func() {
		minorityLock.Lock()
		// Equivalent to Put(cfg, minorityClerkA, "1", "15")
		fd2 := fs.HelpOpen(t, minorityClerkA, dataFile, fs.ReadWrite, 0)
		fs.HelpSeek(t, minorityClerkA, fd2, 0, fs.FromBeginning)
		fs.HelpWriteBytes(t, minorityClerkA, fd2, []byte("15")) //offset now 2
		fs.HelpClose(t, minorityClerkA, fd2)
		minorityLock.Unlock()
		doneWithWriteInMinority <- true
	}()
	go func() {
		minorityLock.Lock()
		// Equivalent to Get(cfg, minorityClerkB, "1") // different clerk in minority
		// Perhaps Get => fs.Seek, fs.Read
		fd2 := fs.HelpOpen(t, minorityClerkB, dataFile, fs.ReadOnly, 0)
		fs.HelpSeek(t, minorityClerkB, fd2, 0, fs.FromBeginning)
		_, _ = fs.HelpRead(t, minorityClerkB, fd2, 2)
		// not sure what the data should be here yet, original test
		// just does the Get and doesn't look at data
		//fs.HelpVerifyBytes(t, dataCkp2b, []byte("??"), "clerkMajority read 1.5")
		fs.HelpClose(t, minorityClerkB, fd2)
		minorityLock.Unlock()
		doneWithReadInMinority <- true
	}()

	select {
	case <-doneWithWriteInMinority:
		t.Fatalf("Write in minority completed before network partition was healed!")
	case <-doneWithReadInMinority:
		t.Fatalf("Read in minority completed before network partition was healed!")
	case <-time.After(time.Second):
	}

	// Make sure that the state has not been changed by requests made to the minority.
	//Equivalent to check(cfg, t, clerkMajority, "1", "14")
	fd = fs.HelpOpen(t, majorityClerk, dataFile, fs.ReadOnly, 0)
	fs.HelpSeek(t, majorityClerk, fd, 0, fs.FromBeginning)
	_, data = fs.HelpRead(t, majorityClerk, fd, 2) //offset now 2
	fs.HelpVerifyBytes(t, data, []byte("14"), "clerkMajority read two")
	fs.HelpClose(t, majorityClerk, fd)

	// Change the state to make sure we can still change it in the majority.
	// Eqivalent to Put(cfg, clerkMajority, "1", "16")
	fd = fs.HelpOpen(t, majorityClerk, dataFile, fs.ReadWrite, 0)
	fs.HelpSeek(t, majorityClerk, fd, 0, fs.FromBeginning)
	fs.HelpWriteBytes(t, majorityClerk, fd, []byte("16")) //offset now 2

	//Equivalent to check(cfg, t, clerkMajority, "1", "16")
	fs.HelpSeek(t, majorityClerk, fd, 0, fs.FromBeginning)
	_, data = fs.HelpRead(t, majorityClerk, fd, 2) //offset now 2
	fs.HelpVerifyBytes(t, data, []byte("16"), "clerkMajority read three")
	fs.HelpClose(t, majorityClerk, fd)

	// This followed, not sure what this checks in original test
	//check(cfg, t, clerkMajority, "1", "14") //check MemoryFS state

	cfg.end()

	cfg.begin("Test: completion after heal")

	cfg.ConnectAll()
	cfg.ConnectClient(minorityClerkA, cfg.All())
	cfg.ConnectClient(minorityClerkB, cfg.All())

	time.Sleep(electionTimeout)

	select {
	case <-doneWithWriteInMinority:
	case <-time.After(30 * 100 * time.Millisecond):
		t.Fatalf("Write did not complete")
	}

	select {
	case <-doneWithReadInMinority:
	case <-time.After(30 * 100 * time.Millisecond):
		t.Fatalf("Read did not complete")
	default:
	}

	//Equivalent to check(cfg, t, clerk, "1", "15")
	fd = fs.HelpOpen(t, clerk, dataFile, fs.ReadOnly, 0)
	fs.HelpSeek(t, clerk, fd, 0, fs.FromBeginning)
	_, dataCk := fs.HelpRead(t, clerk, fd, 2) //offset now 2
	fs.HelpVerifyBytes(t, dataCk, []byte("15"), "clerkMajority read three")
	fs.HelpClose(t, clerk, fd)

	cfg.end()
	return
}
