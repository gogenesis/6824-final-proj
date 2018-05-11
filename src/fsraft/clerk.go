package fsraft

import (
	"ad"
	"crypto/rand"
	crand "crypto/rand"
	"filesystem"
	"fmt"
	"labrpc"
	"math/big"
	mrand "math/rand"
	"sync"
	"time"
)

type Clerk struct {
	lock          sync.Mutex
	servers       []*labrpc.ClientEnd
	id            int64 // a unique serial number for this Clerk
	lastLeader    int   // which server was the leader most recently. -1 initially.
	numOperations int   // how many operations this clerk has submitted (including the current one, if one is in progress)
}

func MakeFsClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.lock.Lock()
	ck.servers = servers
	ck.id = nrand()
	ck.lastLeader = mrand.Intn(len(servers))
	ck.numOperations = 0
	ck.lock.Unlock()

	return ck
}

// See the spec for FileSystem::Mkdir.
func (ck *Clerk) Mkdir(path string) (success bool, err error) {
	ab := AbstractOperation{OpType: MkdirOp}
	ab.Path = path

	returnVal := ck.Operation(ab)

	return castMkdirReply(returnVal)
}

// See the spec for FileSystem::Open.
func (ck *Clerk) Open(path string, mode filesystem.OpenMode, flags filesystem.OpenFlags) (fileDescriptor int, err error) {
	ab := AbstractOperation{OpType: OpenOp}
	ab.Path = path
	ab.OpenMode = mode
	if !filesystem.FlagIsSet(flags, filesystem.Block) {
		// Do the operation as normal
		ab.OpenFlags = flags
		returnVal := ck.Operation(ab)
		return castOpenReply(returnVal)
	} else {
		ad.DebugObj(ck, ad.RPC, "Received Blocking open for %v, "+
			"removing the block flag and trying until it succeeds.", path)
		// sanity check this? I am afraid it may be changing other flags accidentally?
		flagsWithoutBlock := flags ^ filesystem.Block
		ad.Assert(!filesystem.FlagIsSet(flagsWithoutBlock, filesystem.Block))
		for {
			fileDescriptor, err = ck.Open(path, mode, flagsWithoutBlock)
			ad.DebugObj(ck, ad.RPC, "Trying Open(%v) again.", path)
			if err != filesystem.AlreadyOpen {
				return
			}
			time.Sleep(100 * time.Millisecond)
			// and try again
		}
	}
}

// See the spec for FileSystem::Close.
func (ck *Clerk) Close(fileDescriptor int) (success bool, err error) {
	ab := AbstractOperation{OpType: CloseOp}
	ab.FileDescriptor = fileDescriptor

	returnVal := ck.Operation(ab)

	return castCloseReply(returnVal)
}

// See the spec for FileSystem::Seek.
func (ck *Clerk) Seek(fileDescriptor int, offset int, base filesystem.SeekMode) (newPosition int, err error) {
	ab := AbstractOperation{OpType: SeekOp}
	ab.FileDescriptor = fileDescriptor
	ab.Offset = offset
	ab.Base = base

	returnVal := ck.Operation(ab)

	return castSeekReply(returnVal)
}

// See the spec for FileSystem::Read.
func (ck *Clerk) Read(fileDescriptor int, numBytes int) (bytesRead int, data []byte, err error) {
	ab := AbstractOperation{OpType: ReadOp}
	ab.FileDescriptor = fileDescriptor
	ab.NumBytes = numBytes

	returnVal := ck.Operation(ab)

	return castReadReply(returnVal)
}

// See the spec for FileSystem::Write.
func (ck *Clerk) Write(fileDescriptor int, numBytes int, data []byte) (bytesWritten int, err error) {
	ab := AbstractOperation{OpType: WriteOp}
	ab.FileDescriptor = fileDescriptor
	ab.NumBytes = numBytes
	ab.Data = data

	returnVal := ck.Operation(ab)

	return castWriteReply(returnVal)
}

// See the spec for FileSystem::Delete.
func (ck *Clerk) Delete(path string) (success bool, err error) {
	ab := AbstractOperation{OpType: DeleteOp}
	ab.Path = path

	returnVal := ck.Operation(ab)

	return castDeleteReply(returnVal)
}

// Perform some operation.
//
// abstractOperation is the operation to be performed, defined in ops.go.
// Returns an []interface{} of appropriate length and types (see filesystem.go).
func (ck *Clerk) Operation(abstractOperation AbstractOperation) []interface{} {
	ck.lock.Lock()
	defer ck.lock.Unlock()
	ck.numOperations++

	ad.DebugObj(ck, ad.RPC, "Beginning %v", abstractOperation.String())
	args := OperationArgs{abstractOperation, ck.id, ck.numOperations, time.Now().UnixNano()}

	// first, try the last leader
	serverToTry := ck.lastLeader
	reply := ck.sendOperation(args, serverToTry)

	for reply.Status != OK {
		serverToTry = (serverToTry + 1) % len(ck.servers)
		//if reply.Status != NotLeader {
		//	ad.DebugObj(ck, ad.RPC, "%v failed with error status %q so trying another server",
		//		abstractOperation.String(), reply.Status.String())
		//}
		time.Sleep(20 * time.Millisecond)
		reply = ck.sendOperation(args, serverToTry)
	}
	ad.AssertEquals(OK, reply.Status)
	ck.lastLeader = serverToTry
	assertReplyTypesValid(abstractOperation.OpType, reply.ReturnValue)
	ad.DebugObj(ck, ad.RPC, "Returning \"%+v\" from %v", reply.ReturnValue, abstractOperation.String())
	return reply.ReturnValue
}

// Send an individual RPC and wait for its response. Lock OUTSIDE this function.
func (ck *Clerk) sendOperation(args OperationArgs, serverNum int) OperationReply {
	reply := OperationReply{}
	argsCopy := args // make a copy to avoid passing around one object that could be changed. Might be unnecessary?
	// can happen in a lock because clerks only do one request at a time
	//ad.DebugObj(ck, ad.TRACE, "Sending %+v", argsCopy)
	ck.servers[serverNum].Call("FileServer.Operation", &argsCopy, &reply)
	//ad.DebugObj(ck, ad.TRACE, "got %+v in response to %+v", reply, argsCopy)
	return reply
}

// Compress a clerk's int64 ID into something easier to read.
func clerkShortName(clerkId int64) string {
	return fmt.Sprintf("C%03d", clerkId%1000)
}

// Display the status of a clerk: its ID and number of operations.
func (ck *Clerk) DebugPrefix() string {
	return fmt.Sprintf("%v %d", clerkShortName(ck.id), ck.numOperations)
}

// Generate a random integer to use as this clerk's ID.
func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(crand.Reader, max)
	x := bigx.Int64()
	return x
}
