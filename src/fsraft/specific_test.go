package fsraft

import (
	fs "filesystem"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
   "linearizability"
   "bytes"
   "strconv"
   "time"
   "math/rand"
   "sync"
   "sync/atomic"
   "fmt"
)

const electionTimeout = 1 * time.Second
const linearizabilityCheckTimeout = 1 * time.Second

// Specific tests ======================================================================================================
func TestOneClerkFiveServersPartition(t *testing.T) {
	const nservers = 5
	cfg := make_config(t, nservers, false, -1)
	defer cfg.cleanup()
	clerk := cfg.makeClerk(cfg.All())
	dataFile := "/myFile.txt"

	// Equivalent to Put(cfg, clerk, "1", "13")
	fd := fs.HelpOpen(t, clerk, dataFile, fs.ReadWrite, fs.Create)
	fs.HelpWriteBytes(t, clerk, fd, []byte("13")) //offset now 2
	fs.HelpClose(t, clerk, fd)

	cfg.begin("Test: progress in majority")

	majority, minority := cfg.make_partition()
	cfg.partition(majority, minority)

	majorityClerk := cfg.makeClerk(majority)
	minorityClerkA := cfg.makeClerk(minority)
	minorityClerkB := cfg.makeClerk(minority)

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
	go func() {
		// Equivalent to Put(cfg, minorityClerkA, "1", "15")
		fd2 := fs.HelpOpen(t, minorityClerkA, dataFile, fs.ReadWrite, fs.Block)
		fs.HelpSeek(t, minorityClerkA, fd2, 0, fs.FromBeginning)
		fs.HelpWriteBytes(t, minorityClerkA, fd2, []byte("15")) //offset now 2
		fs.HelpClose(t, minorityClerkA, fd2)
		doneWithWriteInMinority <- true
	}()
	go func() {
		// Equivalent to Get(cfg, minorityClerkB, "1") // different clerk in minority
		// Perhaps Get => fs.Seek, fs.Read
		fd2 := fs.HelpOpen(t, minorityClerkB, dataFile, fs.ReadOnly, fs.Block)
		fs.HelpSeek(t, minorityClerkB, fd2, 0, fs.FromBeginning)
		_, _ = fs.HelpRead(t, minorityClerkB, fd2, 2)
		// not sure what the data should be here yet, original test
		// just does the Get and doesn't look at data
		//fs.HelpVerifyBytes(t, dataCkp2b, []byte("??"), "clerkMajority read 1.5")
		fs.HelpClose(t, minorityClerkB, fd2)
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
	fd = fs.HelpOpen(t, majorityClerk, dataFile, fs.ReadOnly, fs.Block)
	fs.HelpSeek(t, majorityClerk, fd, 0, fs.FromBeginning)
	_, data = fs.HelpRead(t, majorityClerk, fd, 2) //offset now 2
	fs.HelpVerifyBytes(t, data, []byte("14"), "clerkMajority read two")
	fs.HelpClose(t, majorityClerk, fd)

	// Change the state to make sure we can still change it in the majority.
	// Eqivalent to Put(cfg, clerkMajority, "1", "16")
	fd = fs.HelpOpen(t, majorityClerk, dataFile, fs.ReadWrite, fs.Block)
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
	cfg.ConnectClerk(minorityClerkA, cfg.All())
	cfg.ConnectClerk(minorityClerkB, cfg.All())

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

// Generic test apparatus =======================================================================================================

// Generic test apparatus ==============================================================================================
// Atomic get/put/append, key is the filename including the "/" and values
// are the entire file contents.
func Get(t *testing.T, clerk *Clerk, fileName string) string {
	fileDescriptor := fs.HelpOpen(t, clerk, fileName, fs.ReadOnly, fs.Create|fs.Block)
	fileLength := fs.HelpSeek(t, clerk, fileDescriptor, 0, fs.FromEnd)
	fs.HelpSeek(t, clerk, fileDescriptor, 0, fs.FromBeginning)
	_, dataBytes := fs.HelpRead(t, clerk, fileDescriptor, fileLength)
	fs.HelpClose(t, clerk, fileDescriptor)
	return string(dataBytes)
}

func Put(t *testing.T, clerk *Clerk, fileName string, newContents string) {
	fileDescriptor := fs.HelpOpen(t, clerk, fileName, fs.WriteOnly, fs.Create|fs.Truncate|fs.Block)
	fs.HelpWriteString(t, clerk, fileDescriptor, newContents)
	fs.HelpClose(t, clerk, fileDescriptor)
}

func Append(t *testing.T, clerk *Clerk, fileName string, value string) {
	fileDescriptor := fs.HelpOpen(t, clerk, fileName, fs.WriteOnly, fs.Block)
	fs.HelpSeek(t, clerk, fileDescriptor, 0, fs.FromEnd)
	fs.HelpWriteString(t, clerk, fileDescriptor, value)
	fs.HelpClose(t, clerk, fileDescriptor)
}

// Basic test is as follows: one or more clerks submitting put/get/append
// operations to set of servers for some period of time.  After the period is
// over, test checks that all appended values are present and in order for a
// particular key.  If unreliable is set, RPCs may fail.  If crash is set, the
// servers crash after the period is over and restart.  If partitions is set,
// the test repartitions the network concurrently with the clerks and servers. If
// maxraftstate is a positive number, the size of the state for Raft (i.e., log
// size) shouldn't exceed 2*maxraftstate.
func GenericTest(t *testing.T, nClerks int, unreliable bool, crash bool, partitions bool, maxraftstate int) {

	title := "Test: "
	if unreliable {
		// the network drops RPC requests and replies.
		title = title + "unreliable net, "
	}
	if crash {
		// peers re-start, and thus persistence must work.
		title = title + "restarts, "
	}
	if partitions {
		// the network may partition
		title = title + "partitions, "
	}
	if maxraftstate != -1 {
		title = title + "snapshots, "
	}
	if nClerks > 1 {
		title = title + "many clerks"
	} else {
		title = title + "one clerk"
	}

	const nservers = 5
	cfg := make_config(t, nservers, unreliable, maxraftstate)
	defer cfg.cleanup()

	cfg.begin(title)
	fmt.Println(title)

	clerkConnectedToAll := cfg.makeClerk(cfg.All())

	partitionerShouldStop := int32(0)
	clerksShouldStop := int32(0)
	partitionerDoneCh := make(chan bool)
	clerkChannels := make([]chan int, nClerks)
	for i := 0; i < nClerks; i++ {
		clerkChannels[i] = make(chan int)
	}

	for i := 0; i < 3; i++ {
		// log.Printf("Iteration %v\n", clerkNum)
		atomic.StoreInt32(&clerksShouldStop, 0)
		atomic.StoreInt32(&partitionerShouldStop, 0)
		go spawn_clerks_and_wait(t, cfg, nClerks, func(clerkNum int, myClerk *Clerk, t *testing.T) {
			numWrites := 0
			defer func() {
				clerkChannels[clerkNum] <- numWrites
			}()
			//fileName := "/" + strconv.Itoa(clerkNum) + ".txt"
			fileName := getFileName(clerkNum)
			expectedContents := ""
			Put(t, myClerk, fileName, expectedContents)
			for atomic.LoadInt32(&clerksShouldStop) == 0 {
				if (rand.Int() % 1000) < 500 {
					toAppend := makeValue(clerkNum, numWrites)
					Append(t, myClerk, fileName, toAppend)
					expectedContents += toAppend
					numWrites++
				} else {
					fileContents := Get(t, myClerk, fileName)
					if fileContents != expectedContents {
						t.Fatalf("get wrong value, fileName %v, wanted:\n\"%v\"\n, "+
							"got\n\"%v\"\n", fileName, expectedContents, fileContents)
					}
				}
			}
		})

		if partitions {
			// Allow the clerks to perform some operations without interruption
			time.Sleep(1 * time.Second)
			go partitioner(t, cfg, partitionerDoneCh, &partitionerShouldStop)
		}
		time.Sleep(5 * time.Second)

		atomic.StoreInt32(&clerksShouldStop, 1)      // tell clerks to quit
		atomic.StoreInt32(&partitionerShouldStop, 1) // tell partitioner to quit

		if partitions {
			// log.Printf("wait for partitioner\n")
			<-partitionerDoneCh
			// reconnect network and submit a request. A clerk may
			// have submitted a request in a minority.  That request
			// won't return until that server discovers a new term
			// has started.
			cfg.ConnectAll()
			// wait for a while so that we have a new term
			time.Sleep(electionTimeout)
		}

		if crash {
			// log.Printf("shutdown servers\n")
			for i := 0; i < nservers; i++ {
				cfg.ShutdownServer(i)
			}
			// Wait for a while for servers to shutdown, since
			// shutdown isn't a real crash and isn't instantaneous
			time.Sleep(electionTimeout)
			// log.Printf("restart servers\n")
			// crash and re-start all
			for i := 0; i < nservers; i++ {
				cfg.StartServer(i)
			}
			cfg.ConnectAll()
		}

		log.Printf("wait for clerks\n")
		for clerkNum := 0; clerkNum < nClerks; clerkNum++ {
			log.Printf("read from clerks %d\n", clerkNum)
			numWrites := <-clerkChannels[clerkNum]
			key := getFileName(clerkNum)
			log.Printf("Check %v writes from clerk %d\n", numWrites, clerkNum)
			// Make sure that the contents of that clerk's file are correct
			fileContents := Get(t, clerkConnectedToAll, key)
			checkClerkAppends(t, clerkNum, fileContents, numWrites)
		}

		if maxraftstate > 0 {
			// Check maximum after the servers have processed all clerk
			// requests and had time to checkpoint.
			if cfg.LogSize() > 2*maxraftstate {
				t.Fatalf("logs were not trimmed (%v > 2*%v)", cfg.LogSize(), maxraftstate)
			}
		}
	}

	cfg.end()
}

// spawn ncli clerks and wait until they are all done
func spawn_clerks_and_wait(t *testing.T, cfg *config, nClerks int, fn func(clerkNum int, ck *Clerk, t *testing.T)) {
	ca := make([]chan bool, nClerks)
	for cli := 0; cli < nClerks; cli++ {
		ca[cli] = make(chan bool)
		go run_clerk(t, cfg, cli, ca[cli], fn)
	}
	log.Printf("spawn_clerks_and_wait: waiting for clerks")
	for cli := 0; cli < nClerks; cli++ {
		ok := <-ca[cli]
		log.Printf("spawn_clerks_and_wait: clerk %d is done\n", cli)
		if ok == false {
			t.Fatalf("failure")
		}
	}
}

// a clerk runs the function f and then signals it is done
func run_clerk(t *testing.T, cfg *config, me int, ca chan bool, fn func(me int, ck *Clerk, t *testing.T)) {
	ok := false
	defer func() { ca <- ok }()
	ck := cfg.makeClerk(cfg.All())
	fn(me, ck, t)
	ok = true
	cfg.deleteClerk(ck)
}

// Get a convenient value that will be written to a file and is easy to debug.
func makeValue(clerkNum int, writeNum int) string {
	return fmt.Sprintf("(%d, %d)", clerkNum, writeNum)
}

// repartition the servers periodically
func partitioner(t *testing.T, cfg *config, doneChannel chan bool, shouldStop *int32) {
	defer func() { doneChannel <- true }()
	for atomic.LoadInt32(shouldStop) == 0 {
		array := make([]int, cfg.n)
		for i := 0; i < cfg.n; i++ {
			array[i] = (rand.Int() % 2)
		}
		partition := make([][]int, 2)
		for halfNum := 0; halfNum < 2; halfNum++ {
			partition[halfNum] = make([]int, 0)
			for serverNum := 0; serverNum < cfg.n; serverNum++ {
				if array[serverNum] == halfNum {
					partition[halfNum] = append(partition[halfNum], serverNum)
				}
			}
		}
		cfg.partition(partition[0], partition[1])
		time.Sleep(electionTimeout + time.Duration(rand.Int63()%200)*time.Millisecond)
	}
}

// Get the name of the file that clerk number clerkNum will write to.
func getFileName(clerkNum int) string {
	return "/" + strconv.Itoa(clerkNum) + ".txt"
}

// check that for a specific clerk all known appends are present in a value,
// and in order
func checkClerkAppends(t *testing.T, clerkNum int, fileContents string, count int) {
	lastoff := -1
	for writeNum := 0; writeNum < count; writeNum++ {
		wanted := makeValue(clerkNum, writeNum)
		off := strings.Index(fileContents, wanted)
		if off < 0 {
			t.Fatalf("%v missing element %v in Append result %v", clerkNum, wanted, fileContents)
		}
		off1 := strings.LastIndex(fileContents, wanted)
		if off1 != off {
			t.Fatalf("duplicate element %v in Append result", wanted)
		}
		if off <= lastoff {
			t.Fatalf("wrong order for element %v in Append result", wanted)
		}
		lastoff = off
	}
}

// Generic tests =======================================================================================================

func TestBasicKV(t *testing.T) {
	// Basically use the filesystem as a key-value store
	GenericTest(t, 1, false, false, false, -1)
}

func TestConcurrentKV(t *testing.T) {
	GenericTest(t, 5, false, false, false, -1)
}

func TestUnreliableKV(t *testing.T) {
	GenericTest(t, 5, true, false, false, -1)
}

func TestManyPartitionsOneClientKV(t *testing.T) {
	GenericTest(t, 1, false, false, true, -1)
}

func TestManyPartitionsManyClientsKV(t *testing.T) {
	GenericTest(t, 5, false, false, true, -1)
}

func TestPersistOneClientKV(t *testing.T) {
	GenericTest(t, 1, false, true, false, -1)
}

func TestPersistConcurrentKV(t *testing.T) {
	GenericTest(t, 5, false, true, false, -1)
}

func TestPersistConcurrentUnreliableKV(t *testing.T) {
	GenericTest(t, 5, true, true, false, -1)
}

func TestPersistPartitionKV(t *testing.T) {
	GenericTest(t, 5, false, true, true, -1)
}

func TestPersistPartitionUnreliableKV(t *testing.T) {
	GenericTest(t, 5, true, true, true, -1)
}

func TestSnapshotRecoverKV(t *testing.T) {
	GenericTest(t, 1, false, true, false, 1000)
}

func TestSnapshotRecoverManyClientsKV(t *testing.T) {
	GenericTest(t, 20, false, true, false, 1000)
}

func TestSnapshotUnreliableKV(t *testing.T) {
	GenericTest(t, 5, true, false, false, 1000)
}

func TestSnapshotUnreliableRecoverKV(t *testing.T) {
	GenericTest(t, 5, true, true, false, 1000)
}

func TestSnapshotUnreliableRecoverConcurrentPartitionKV(t *testing.T) {
	GenericTest(t, 5, true, true, true, 1000)
}

// a client runs the function f and then signals it is done
func run_client(t *testing.T, cfg *config, me int, ca chan bool, fn func(me int, ck *Clerk, t *testing.T)) {
	ok := false
	defer func() { ca <- ok }()
	ck := cfg.makeClient(cfg.All())
	fn(me, ck, t)
	ok = true
	cfg.deleteClient(ck)
}



// spawn ncli clients and wait until they are all done
func spawn_clients_and_wait(t *testing.T, cfg *config, ncli int, fn func(me int, ck *Clerk, t *testing.T)) {
	ca := make([]chan bool, ncli)
	for cli := 0; cli < ncli; cli++ {
		ca[cli] = make(chan bool)
		go run_client(t, cfg, cli, ca[cli], fn)
	}
	// log.Printf("spawn_clients_and_wait: waiting for clients")
	for cli := 0; cli < ncli; cli++ {
		ok := <-ca[cli]
		// log.Printf("spawn_clients_and_wait: client %d is done\n", cli)
		if ok == false {
			t.Fatalf("failure")
		}
	}
}

// similar to GenericTest, but with clients doing random operations (and using a
// linearizability checker)
func GenericTestLinearizability(t *testing.T, part string, nclients int, nservers int, unreliable bool, crash bool, partitions bool, maxraftstate int) {
	title := "Test: "
	if unreliable {
		// the network drops RPC requests and replies.
		title = title + "unreliable net, "
	}
	if crash {
		// peers re-start, and thus persistence must work.
		title = title + "restarts, "
	}
	if partitions {
		// the network may partition
		title = title + "partitions, "
	}
	if maxraftstate != -1 {
		title = title + "snapshots, "
	}
	if nclients > 1 {
		title = title + "many clients"
	} else {
		title = title + "one client"
	}
	title = title + ", linearizability checks (" + part + ")" // 3A or 3B

	cfg := make_config(t, nservers, unreliable, maxraftstate)
	defer cfg.cleanup()

	cfg.begin(title)

	begin := time.Now()
	var operations []linearizability.Operation
	var opMu sync.Mutex

	done_partitioner := int32(0)
	done_clients := int32(0)
	ch_partitioner := make(chan bool)
	clnts := make([]chan int, nclients)
	for i := 0; i < nclients; i++ {
		clnts[i] = make(chan int)
	}
	for i := 0; i < 3; i++ {
		// log.Printf("Iteration %v\n", i)
		atomic.StoreInt32(&done_clients, 0)
		atomic.StoreInt32(&done_partitioner, 0)
		go spawn_clients_and_wait(t, cfg, nclients, func(cli int, myck *Clerk, t *testing.T) {
			j := 0
			defer func() {
				clnts[cli] <- j
			}()
			for atomic.LoadInt32(&done_clients) == 0 {
				key := fmt.Sprintf("/lin-rnd-%d.txt", strconv.Itoa(rand.Int() % nclients))
				nv := []bytes("x " + strconv.Itoa(cli) + " " + strconv.Itoa(j) + " y")
				var inp linearizability.KvInput  //Write
				var out linearizability.KvOutput //Read
				start := int64(time.Since(begin))
				if (rand.Int() % 1000) < 500 {
					//Equivalent to Append(cfg, myck, key, nv)
               fd := myck.HelpOpen(t, myck, key)
               HelpWrite(t, myck, fd, key, nv)
               HelpClose(t, myck, fd)
					//Equivalent to inp = linearizability.KvInput{Op: 2, Key: key, Value: nv}
					inp = linearizability.MemFSInput{ Op: 2, Key: key, Value: nv }
					j++
				} else if (rand.Int() % 1000) < 100 {
					//Equivalent to Put(cfg, myck, key, nv)
               fd := myck.HelpOpen(t, myck, key)
               HelpSeek(t, myck, fd, 0, FromBeginning)
               HelpWrite(t, myck, key, nv)
               HelpClose(t, myck, fd)
					//Equivalent to inp = linearizability.KvInput{Op: 1, Key: key, Value: nv}
					inp = linearizability.MemFSInput{Op: 1, Key: key, Value: nv}
					j++
				} else {
					//Equivalent to v := Get(cfg, myck, key)
               fd := myck.HelpOpen(t, myck, key)
               HelpSeek(t, myck, fd, 0, FromBeginning)
               v := HelpRead(t, myck, fd, len(nv))
               HelpClose(t, myck, fd)

					//Eqiv. to inp = linearizability.KvInput{Op: 0, Key: key}
					inp = linearizability.MemFSInput{Op: 0, Key: key}
					//Eqiv. to out = linearizability.KvOutput{Value: v}
					out = linearizability.MemFSOutput{Value: v}
				}
				end := int64(time.Since(begin))
				op := linearizability.Operation{Input: inp, Call: start, Output: out, Return: end}
				opMu.Lock()
				operations = append(operations, op)
				opMu.Unlock()
			}
		})

		if partitions {
			// Allow the clients to perform some operations without interruption
			time.Sleep(1 * time.Second)
			go partitioner(t, cfg, ch_partitioner, &done_partitioner)
		}
		time.Sleep(5 * time.Second)

		atomic.StoreInt32(&done_clients, 1)     // tell clients to quit
		atomic.StoreInt32(&done_partitioner, 1) // tell partitioner to quit

		if partitions {
			// log.Printf("wait for partitioner\n")
			<-ch_partitioner
			// reconnect network and submit a request. A client may
			// have submitted a request in a minority.  That request
			// won't return until that server discovers a new term
			// has started.
			cfg.ConnectAll()
			// wait for a while so that we have a new term
			time.Sleep(electionTimeout)
		}

		if crash {
			// log.Printf("shutdown servers\n")
			for i := 0; i < nservers; i++ {
				cfg.ShutdownServer(i)
			}
			// Wait for a while for servers to shutdown, since
			// shutdown isn't a real crash and isn't instantaneous
			time.Sleep(electionTimeout)
			// log.Printf("restart servers\n")
			// crash and re-start all
			for i := 0; i < nservers; i++ {
				cfg.StartServer(i)
			}
			cfg.ConnectAll()
		}

		// wait for clients.
		for i := 0; i < nclients; i++ {
			<-clnts[i]
		}

		if maxraftstate > 0 {
			// Check maximum after the servers have processed all client
			// requests and had time to checkpoint.
			if cfg.LogSize() > 2*maxraftstate {
				t.Fatalf("logs were not trimmed (%v > 2*%v)", cfg.LogSize(), maxraftstate)
			}
		}
	}

	cfg.end()

	// log.Printf("Checking linearizability of %d operations", len(operations))
	// start := time.Now()
	//Equiv to ok :=  ... KvModel(), operations, linearizabilityCheckTimeout)
	ok := linearizability.CheckOperationsTimeout(linearizability.MemFSModel(), operations, linearizabilityCheckTimeout)
	// dur := time.Since(start)
	// log.Printf("Linearizability check done in %s; result: %t", time.Since(start).String(), ok)
	if !ok {
		t.Fatal("history is not linearizable")
	}
}

func DISABLED_PersistPartitionUnreliableLinearizable3A(t *testing.T) {
   // Test: unreliable net, restarts, partitions, linearizability checks (3A) ...
   GenericTestLinearizability(t, "3A", 15, 7, true, true, true, -1)
}
