package fsraft

import (
	"ad"
	"bytes"
	"filesystem"
	"fmt"
	"labgob"
	"labrpc"
	"memoryFS"
	"raft"
	"sync"
	"time"
)

// A file server built on Raft.
// This is similar to the kvserver, but stores an in-memory FileSystem (see package "memoryFS")
// instead of a map[string]string. This filesystem is linearizable; a Read() that
// begins after a Write() finishes is guaranteed to return the new data.
type FileServer struct {
	lock               sync.Mutex
	me                 int                // index into the list of servers
	applyCh            chan raft.ApplyMsg // for messages from the raft server
	rf                 *raft.Raft         // The underlying Raft
	maxraftstate       int                // snapshot if log grows this big, -1 for no snapshots
	killCh             chan bool          // send on this channel when you die
	thinksRaftIsLeader bool               // if it thinks its raft peer is a leader
	thinksRaftTermIs   int                // what it thinks the term of the underlying Raft peer is

	memoryFS                 memoryFS.MemoryFS // The actual filesystem stored on this server
	operationsInProgress     map[*OperationArgs]OperationInProgress
	clerkCommandsExecuted    map[int64]int //clerkCommandsExecuted[clerk serial number] = last command index of a command from that clerk
	lastCommandIndexExecuted int           // total number of commands executed. Equal to the sum of values in clerkCommandsExecuted.
}

// Data written must be arrays of this length instead of slices.
// This is so that AbstractOperations can be compared with ==, used as map keys, etc,
// all of which can be done with arrays but not not slices.
const WriteSizeBytes = 64

// Clerk-facing API ====================================================================================================

// Start a FileServer.
// servers[] contains the ports of the set of servers that will cooperate via Raft to form the fault-tolerant file service.
// me is the index of the current server in servers[].
func StartFileServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxRaftState int) *FileServer {
	// the filesystem server should store snapshots with persister.SaveSnapshot(),
	// and Raft should save its state (including log) with persister.SaveRaftState().
	// the filesystem server should snapshot when Raft's saved state exceeds maxraftstate bytes,
	// in order to allow Raft to garbage-collect its log. if maxraftstate is -1, you don't need to snapshot.
	// StartFileServer() must return quickly, so it should start goroutines
	// for any long-running work.

	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(OpenOp)
	labgob.Register(filesystem.NotFound)
	labgob.Register(filesystem.ReadOnly)
	labgob.Register(filesystem.Append)
	labgob.Register(filesystem.FromBeginning)
	labgob.Register(AbstractOperation{})
	labgob.Register(OperationArgs{})
	labgob.Register(OperationReply{})

	fs := new(FileServer)
	fs.lock.Lock()
	fs.me = me
	fs.applyCh = make(chan raft.ApplyMsg)
	fs.rf = raft.Make(servers, me, persister, fs.applyCh)
	fs.maxraftstate = maxRaftState
	fs.killCh = make(chan bool, 2) // 2 because there's 2 long-running threads per server

	fs.thinksRaftIsLeader = false
	fs.thinksRaftTermIs = 0

	fs.memoryFS = memoryFS.CreateEmptyMemoryFS()
	fs.operationsInProgress = make(map[*OperationArgs]OperationInProgress)
	fs.clerkCommandsExecuted = make(map[int64]int)
	fs.lastCommandIndexExecuted = 0

	go fs.applyChMonitorThread()
	go fs.stateUpdaterThread()

	fs.readSnapshot(persister.ReadSnapshot())

	fs.lock.Unlock()
	return fs
}

// Kill a FileServer.
func (fs *FileServer) Kill() {
	fs.lock.Lock()
	ad.DebugObj(fs, ad.TRACE, "Dying")

	// send two times because there are two long-running threads per kvServer
	fs.killCh <- true
	fs.killCh <- true

	fs.rf.Kill()
	fs.lock.Unlock()
}

// Do an Operation.
func (fs *FileServer) Operation(args *OperationArgs, reply *OperationReply) {
	fs.lock.Lock()
	ad.Assert(args != nil)
	ad.Assert(reply != nil)

	expectedIndex, startTerm, isLeader := fs.rf.Start(*args)
	fs.updateTermAndLeadershipToValues(startTerm, isLeader)
	if !isLeader {
		reply.Status = NotLeader
		fs.lock.Unlock()
		return
	}

	// unbuffered because I'll be waiting whenever someone sends on this channel
	resultChannel := make(chan OperationReply)
	fs.operationsInProgress[args] = OperationInProgress{*args, expectedIndex, resultChannel}
	ad.DebugObj(fs, ad.RPC, "Started %v for %v %d", args.AbstractOperation.String(), clerkShortName(args.ClerkId), args.ClerkIndex)
	fs.lock.Unlock()

	result := <-resultChannel

	fs.lock.Lock()
	defer fs.lock.Unlock()

	// Copy fields because the OperationReply we get out of the channel is a // different object than
	// the OperationReply the client is expecting to get mutated.
	reply.Status = result.Status
	reply.ReturnValue = result.ReturnValue
}

// Long-running threads ================================================================================================

func (fs *FileServer) applyChMonitorThread() {
	// This thread monitors applyCh and waits for things to come out of it.
	for {
	waitForApplyMsgs:
		select {
		case <-fs.killCh:
			fs.lock.Lock()
			for _, opInProgress := range fs.operationsInProgress {
				go func(opInprogress OperationInProgress) {
					// that's an empty List<Object> but in Go it's []interface{}{}
					opInProgress.resultChannel <- OperationReply{[]interface{}{}, Killed}
				}(opInProgress)
			}
			fs.lock.Unlock()
			return

		case applyMsg := <-fs.applyCh:
			fs.lock.Lock()
			ad.Assert(applyMsg.CommandValid)
			ad.DebugObj(fs, ad.TRACE, "Got %+v out of the ApplyCh", applyMsg)
			if applyMsg.Purpose == raft.COMMAND {
				opArgs := applyMsg.Command.(OperationArgs)
				ad.DebugObj(fs, ad.TRACE, "This is %v %d", clerkShortName(opArgs.ClerkId), opArgs.ClerkIndex)
				fs.updateTermAndLeadership()

				// if this applyMsg is for a client who has an operation in progress with us
				opInProgress, containsKey := fs.operationsInProgress[&opArgs]

				if containsKey {
					// Maybe lose leadership. If a different command has appeared at the index returned by Start
					// equivalently, if one of our calls in progress was supposed to put something at the index we just got
					// (and we got something different there)
					for _, otherOpInProgress := range fs.operationsInProgress {
						if (applyMsg.CommandIndex == otherOpInProgress.expectedIndex) &&
							!OperationArgsEquals(opArgs, otherOpInProgress.operationArgs) {
							ad.DebugObj(fs, ad.WARN, "A different operation has appeared at the index returned by Start()!"+
								" I have %+v in progress, but %+v is different and has the same index.", opInProgress, otherOpInProgress)
							fs.loseLeadership()
							break
						}
					}
				}

				if applyMsg.CommandIndex < fs.lastCommandIndexExecuted+1 {
					ad.DebugObj(fs, ad.WARN, "Skipping out-of-order command %+v!", applyMsg)
					goto waitForApplyMsgs
				}

				returnValue := fs.execute(opArgs.AbstractOperation, opArgs.ClerkId, opArgs.ClerkIndex, applyMsg.CommandIndex)

				if containsKey {
					ad.DebugObj(fs, ad.TRACE, "Routing RPC reply OK to %v %d", clerkShortName(opArgs.ClerkId), opArgs.ClerkIndex)
					opInProgress.resultChannel <- OperationReply{returnValue, OK}
					delete(fs.operationsInProgress, &opArgs)
				} else {
					clientsInProgressShortNames := make([]string, len(fs.operationsInProgress))
					for otherOperationArgs := range fs.operationsInProgress {
						clientName := clerkShortName(otherOperationArgs.ClerkId)
						clientsInProgressShortNames = append(clientsInProgressShortNames, clientName)
					}
					ad.DebugObj(fs, ad.TRACE, "No RPC in progress for %v, clients with RPCs in progress = %+v",
						clerkShortName(opArgs.ClerkId), clientsInProgressShortNames)
				}
				if (fs.maxraftstate != -1) && (fs.rf.StateSizeBytes() > fs.maxraftstate) {
					ad.DebugObj(fs, ad.TRACE, "Raft's state is %d bytes, but max is %d bytes.", fs.rf.StateSizeBytes(), fs.maxraftstate)
					ad.AssertEquals(applyMsg.CommandIndex, fs.lastCommandIndexExecuted)
					fs.writeSnapshot(fs.lastCommandIndexExecuted)
				}
			} else {
				index := applyMsg.CommandIndex
				if index > fs.lastCommandIndexExecuted {
					ad.DebugObj(fs, ad.RPC, "Got a snapshot out of the applyCh! index=%d, term=%d. Applying snapshot.", index, applyMsg.CommandTerm)
					fs.lastCommandIndexExecuted = index
					fs.updateTermAndLeadership()
					fs.readSnapshot(applyMsg.Command.([]byte))
					fs.writeSnapshot(index) // so it can be backed up
				} else {
					ad.DebugObj(fs, ad.TRACE, "Ignoring snapshot out of the applyCh because it covers indices through %d and I have"+
						" already executed commands through index %d.", index, fs.lastCommandIndexExecuted)
					fs.updateTermAndLeadership()
				}
			}
			fs.lock.Unlock()
		} // end select
	} // end for
}

// May not be necessary?
func (fs *FileServer) stateUpdaterThread() {
	for {
		select {
		case <-fs.killCh:
			return
		case <-time.After(300 * time.Millisecond):
			fs.lock.Lock()
			fs.updateTermAndLeadership()
			fs.lock.Unlock()
		}
	}
}

// Private helper methods ==============================================================================================

func (fs *FileServer) execute(ab AbstractOperation, clerkId int64, clerkIndex int, commandIndex int) []interface{} {
	// Don't just return from here because we might execute a duplicate op anyway if it's a GetOp
	isDuplicate := false
	duplicateReason := ""
	if clerkIndex <= fs.clerkCommandsExecuted[clerkId] {
		isDuplicate = true
		duplicateReason = fmt.Sprintf("this is the %dth command from %v and I have already executed %d commands from them",
			clerkIndex, clerkShortName(clerkId), fs.clerkCommandsExecuted[clerkId])
	}
	if commandIndex <= fs.lastCommandIndexExecuted {
		isDuplicate = true
		duplicateReason = fmt.Sprintf("it is at commandIndex=%d and I have already executed %d commands",
			commandIndex, fs.lastCommandIndexExecuted)
	}

	// Perform commands in the right order
	if fs.lastCommandIndexExecuted+1 != commandIndex {
		debugStr := fmt.Sprintf("Executing %v for %v %d, expecting commandIndex=%d, but got %d instead!",
			ab.String(), clerkShortName(clerkId), clerkIndex, fs.lastCommandIndexExecuted+1, commandIndex)
		ad.DebugObj(fs, ad.WARN, debugStr)
		panic(debugStr)
	}
	fs.lastCommandIndexExecuted += 1

	if !isDuplicate {
		// Don't skip commands from a clerk
		ad.AssertEquals(fs.clerkCommandsExecuted[clerkId]+1, clerkIndex)
		fs.clerkCommandsExecuted[clerkId] += 1
	}

	// All methods mutate the memoryFS, so never execute duplicates.
	// (The code is structured like this because in lab 3 the correct behavior was to
	// execute duplicate ops iff they were Gets, and the code is copied from there).
	if !isDuplicate {

		returnValue := fs.performAbstractOperation(ab)
		ad.DebugObj(fs, ad.RPC, "Executing %v for %v %v. State is now %+v", ab.String(), clerkShortName(clerkId),
			clerkIndex, fs.stateStr())
		return returnValue
	} else {
		ad.DebugObj(fs, ad.TRACE, "Skipping duplicate command %+v for %v %d because %v", ab, clerkShortName(clerkId), clerkIndex, duplicateReason)
		return []interface{}{}
	}
}

// Perform an operation on the filesystem and return the result.
func (fs *FileServer) performAbstractOperation(ab AbstractOperation) []interface{} {
	// Should be a switch on OpType
	switch ab.OpType {
	case MkdirOp:
		// Make sure relevant args are set
		ad.Assert(ab.Path != "")
		success, err := fs.memoryFS.Mkdir(ab.Path)
		return []interface{}{success, err}
	case OpenOp:
		ad.Assert(ab.Path != "")
		fileDescriptor, err := fs.memoryFS.Open(ab.Path, ab.OpenMode, ab.OpenFlags)
		return []interface{}{fileDescriptor, err}
	case CloseOp:
		fileDescriptor, err := fs.memoryFS.Close(ab.FileDescriptor)
		return []interface{}{fileDescriptor, err}
	case SeekOp:
		newPosition, err := fs.memoryFS.Seek(ab.FileDescriptor, ab.Offset, ab.Base)
		return []interface{}{newPosition, err}
	case ReadOp:
		bytesRead, data, err := fs.memoryFS.Read(ab.FileDescriptor, ab.NumBytes)
		return []interface{}{bytesRead, data, err}
	case WriteOp:
		// The "[:]" is necessary to convert the fixed-length array into a slice that the memoryfs will accept.
		// Unlike Python, it does not create a copy of the underlying data.
		bytesWritten, err := fs.memoryFS.Write(ab.FileDescriptor, ab.NumBytes, ab.Data[:])
		return []interface{}{bytesWritten, err}
	case DeleteOp:
		success, err := fs.memoryFS.Delete(ab.Path)
		return []interface{}{success, err}
	}
	panic("Needs a return at the end of the function, but we can never get here")
}

// Update what this KVServer thinks its Raft server's term and leadership status are.
// ONLY CALL WITH THE LOCK.
func (fs *FileServer) updateTermAndLeadership() {
	fs.updateTermAndLeadershipToValues(fs.rf.GetState())
}

func (fs *FileServer) updateTermAndLeadershipToValues(actualTerm int, actualIsLeader bool) {
	if actualTerm == fs.thinksRaftTermIs && actualIsLeader == fs.thinksRaftIsLeader {
		// no changes to be made
		return
	}
	ad.DebugObj(fs, ad.TRACE, "Updating term=%d and leader=%v", actualTerm, actualIsLeader)

	// can't go backwards in time
	if actualTerm > fs.thinksRaftTermIs {
		fs.thinksRaftTermIs = actualTerm
	}

	if fs.thinksRaftIsLeader && (!actualIsLeader) {
		fs.loseLeadership() // also sets fs.thinksRaftIsLeader from true to false
	} else if (!fs.thinksRaftIsLeader) && actualIsLeader {
		ad.DebugObj(fs, ad.RPC, "Becoming leader.")
		fs.thinksRaftIsLeader = true
	} else {
		ad.AssertEquals(fs.thinksRaftIsLeader, actualIsLeader)
	}
}

func (fs *FileServer) loseLeadership() {
	ad.Assert(fs.thinksRaftIsLeader)
	if len(fs.operationsInProgress) == 0 {
		ad.DebugObj(fs, ad.WARN, "Lost leadership! Would fail RPCs in progress, but there are none.")
	} else {
		ad.DebugObj(fs, ad.WARN, "Lost leadership! Failing all %d RPCs in progress %+v", len(fs.operationsInProgress), fs.operationsInProgress)
		for clientId, opInProgress := range fs.operationsInProgress {
			// no need to send in separate goroutines because there is guaranteed to be someone waiting on this channel
			opInProgress.resultChannel <- OperationReply{[]interface{}{}, NotLeader}
			delete(fs.operationsInProgress, clientId)
		}
	}
	fs.thinksRaftIsLeader = false
}

// Snapshot methods ====================================================================================================

// Create a snapshot.
// Includes commandIndex up to and including lastIndexInSnapshot
// Not threadsafe: only call with the lock!
func (fs *FileServer) writeSnapshot(lastIncludedIndex int) {
	ad.DebugObj(fs, ad.RPC, "Writing snapshot containing up to index=%d", lastIncludedIndex)
	fs.rf.Snapshot(fs.getSnapshotData(), lastIncludedIndex)

}

// Returns the data to be persisted in a snapshot.
func (fs *FileServer) getSnapshotData() []byte {
	byteBuffer := new(bytes.Buffer)
	encoder := labgob.NewEncoder(byteBuffer)
	encoder.Encode(fs.memoryFS)
	encoder.Encode(fs.clerkCommandsExecuted)
	encoder.Encode(fs.lastCommandIndexExecuted)

	return byteBuffer.Bytes()
}

// Restore this sever's state from a snapshot.
func (fs *FileServer) readSnapshot(data []byte) {
	if data == nil || len(data) == 0 {
		return // can't bootstrap without any state
	}

	byteBuffer := bytes.NewBuffer(data)
	decoder := labgob.NewDecoder(byteBuffer)

	var mfs memoryFS.MemoryFS
	if decoder.Decode(&mfs) != nil {
		panic("Error decoding memoryFS!")
	} else {
		fs.memoryFS = mfs
	}

	var clerkCommandsExecuted map[int64]int
	if decoder.Decode(&clerkCommandsExecuted) != nil {
		panic("Error decoding clerkCommandsExecuted!")
	} else {
		fs.clerkCommandsExecuted = clerkCommandsExecuted
	}

	var lastCommandIndexExecuted int
	if decoder.Decode(&lastCommandIndexExecuted) != nil {
		panic("Error decoding lastCommandIndexExecuted!")
	} else {
		fs.lastCommandIndexExecuted = lastCommandIndexExecuted
	}

	ad.DebugObj(fs, ad.RPC, "State read from stable storage. memoryFS=%+v, clerkCommandsExecuted=%+v, "+
		"lastCommandIndexExecuted=%v", fs.memoryFS, fs.clerkCommandsExecuted, fs.lastCommandIndexExecuted)
}

// Small debugging helper methods ======================================================================================

// Get a pointer to this server's Raft instance. FOR TESTING PURPOSES ONLY.
func (fs *FileServer) Raft() *raft.Raft {
	return fs.rf
}

// Emit a short string with the state for debugging purposes.
func (fs *FileServer) DebugPrefix() string {
	//actualTerm, actualIsLeader := fs.rf.GetState()
	//
	//actualElectionStr := "F"
	//if actualIsLeader {
	//	actualElectionStr = "L"
	//}
	thinksElectionStr := "F"
	if fs.thinksRaftIsLeader {
		thinksElectionStr = "L"
	}

	return fmt.Sprintf("S%v %v%d %d", fs.me, thinksElectionStr, fs.thinksRaftTermIs, fs.lastCommandIndexExecuted)
}

// Get the state as a string for easy debugging
func (fs *FileServer) stateStr() string {
	return "\"TODO\""
}
