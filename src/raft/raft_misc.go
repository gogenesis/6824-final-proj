package raft

import (
	"ad"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"time"
)

func init() {
	// first things first. This is in case you want the same seed for a later test.
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	fmt.Printf("seed=%v\n", seed)
}

// for debugging
func (rf *Raft) DebugPrefix() string {
	var electionState string
	switch rf.CurrentElectionState {
	case Leader:
		electionState = "L"
	case Candidate:
		electionState = "C"
	case Follower:
		electionState = "F"
	}
	return fmt.Sprintf("R%d %v%d %d/%d/%d/%d", rf.me, electionState, rf.CurrentTerm,
		rf.lastIndexInSnapshot(), rf.lastApplied, rf.commitIndex, rf.Log.length())
}

// check representation invariants.
// ONLY CALL WITH THE LOCK
func (rf *Raft) assertInvariants() {
	if !ad.AssertionsEnabled {
		// for performance, don't check invariants
		return
	}
	if !rf.isAlive {
		return
	}
	assertEquals(rf.Log.length(), rf.lastLogIndex())
	if !(0 <= rf.lastIndexInSnapshot() &&
		rf.lastIndexInSnapshot() <= rf.lastApplied &&
		rf.lastApplied <= rf.commitIndex &&
		rf.commitIndex <= rf.lastLogIndex()) {
		panic(fmt.Sprintf("Illegal state: %d/%d/%d/%d", rf.lastIndexInSnapshot(), rf.lastApplied, rf.commitIndex, rf.lastLogIndex()))
	}
	if rf.CurrentElectionState == Leader {
		for peerNum, _ := range rf.peers {
			assert(rf.nextIndex[peerNum] > 0)
			// it could be equal to + 1 if they already have all our entries
			assert(rf.nextIndex[peerNum] <= rf.Log.lastIndex()+1)
			assert(rf.matchIndex[peerNum] >= 0)
			assert(rf.matchIndex[peerNum] <= rf.Log.lastIndex())
		}
	}
}

func (rf *Raft) lastLogIndex() int {
	return rf.Log.lastIndex()
}

func (rf *Raft) majoritySize() int {
	return int(math.Ceil((float64(len(rf.peers) + 1)) / 2))
}

// Returns the last index in the snapshot, or -1 if no snapshot.
func (rf *Raft) lastIndexInSnapshot() int {
	return rf.Log.lastCompressedIndex()
}

// If necesary, updates term to otherTerm. Otherwise, does nothing.
// ONLY CALL WITH THE LOCK
func (rf *Raft) updateTermIfNecessary(otherTerm int) {
	if otherTerm > rf.CurrentTerm {
		rf.CurrentTerm = otherTerm
		rf.VotedFor = -1
		if rf.CurrentElectionState == Leader {
			go func() { rf.becomeFollower <- true }()
		}
		rf.CurrentElectionState = Follower
		ad.DebugObj(rf, ad.RPC, "Updating term to %d and becoming follower", rf.CurrentTerm)
		rf.writePersist()
	}
}

func assert(cond bool) {
	if !cond {
		panic("Assertion failed!")
	}
}

func assertEquals(expected, actual interface{}) {
	if !(expected == actual) {
		panic(fmt.Sprintf("AssertionError: expected %v, got %v\n", expected, actual))
	}
}

func getElectionTimeout() time.Duration {
	ms := minElectionTimeout + rand.Intn(maxElectionTimeout-minElectionTimeout)
	return time.Duration(ms) * time.Millisecond
}

func getHeartbeatTimeout() time.Duration {
	return time.Duration(heartbeatTime) * time.Millisecond
}

func (rf *Raft) resetElectionTimeout() {
	rf.candidateDeclareTime = time.Now().Add(getElectionTimeout())
}

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

const DEBUG_LOCKS = false

func (rf *Raft) lock() {
	var callingFunctionName string
	var lineNum int
	if DEBUG_LOCKS {
		funcptr, _, lineNum, _ := runtime.Caller(1)
		callingFunctionName = runtime.FuncForPC(funcptr).Name()
		fmt.Printf("R%d waiting for lock in %v:%d\n", rf.me, callingFunctionName, lineNum)
	}
	rf.mutex.Lock()
	if DEBUG_LOCKS {
		fmt.Printf("R%d acquired for lock in %v:%d\n", rf.me, callingFunctionName, lineNum)
	}
}

func (rf *Raft) unlock() {
	if DEBUG_LOCKS {
		funcptr, _, lineNum, _ := runtime.Caller(1)
		function := runtime.FuncForPC(funcptr)
		callingFunctionName := function.Name()
		fmt.Printf("R%d releasing for lock in %v:%d\n", rf.me, callingFunctionName, lineNum)
	}
	rf.assertInvariants()
	rf.mutex.Unlock()
}
