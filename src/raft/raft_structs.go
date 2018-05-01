package raft

import (
	"labrpc"
	"sync"
	"time"
)

// peer sends this to the service on the same serve to apply a Log entry
type ApplyMsg struct {
	CommandValid bool // this is currently always true...
	Command      interface{}
	CommandIndex int
	CommandTerm  int
	// If Purpose == COMMAND, then Command is a single command that has just been applied.
	// If Purpose == STATE_RESET, then Command is a byte[] containing the entire state of the state machine.
	Purpose ApplyMsgPurpose
}

type ApplyMsgPurpose int

const (
	COMMAND ApplyMsgPurpose = iota
	STATE_RESET
)

type LogEntry struct {
	Term    int         // the term when the entry was received by the leader
	Command interface{} // The command
	Index   int         // 0-index position in the log
}

// timeouts in milliseconds
const (
	minElectionTimeout = 500
	maxElectionTimeout = 1000
	heartbeatTime      = 150
)

type ElectionState int

const (
	Leader ElectionState = iota
	Candidate
	Follower
)

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	// FINAL: never changed to point to new objects
	mutex          sync.Mutex          // Lock to protect shared access to this peer's state
	peers          []*labrpc.ClientEnd // RPC end points of all peers
	persister      *Persister          // Object to hold this peer's persisted state
	me             int                 // this peer's index into peers[]
	isAlive        bool                // If false, suppresses debug output and stops doing things
	applyCh        chan ApplyMsg       // A channel on which the tester or service expects ApplyMsg messages.
	toApply        chan bool           // send on this channel to apply that entry to the state machine
	becomeLeader   chan int            // broadcast when you become leader. int is the term in which you become leader.
	becomeFollower chan bool           // broadcast when you become not the leader

	// PERSISTENT: always update on stable storage before responding to RPCs
	CurrentTerm          int           // latest term the server has seen, initialized to 0, only increases
	VotedFor             int           // candidateID that I voted for in term CurrentTerm, -1 if none
	Log                  LogOne        // the operations applied to this state machine
	CurrentElectionState ElectionState // Leader, Candidate, or Follower

	// VOLATILE: does not need to be updated before replying to RPCs
	commitIndex          int       // index of highest Log entry known to be committed (initialized to 0, increases monotonically)
	lastApplied          int       //  index of highest Log entry applied to state machine (initialized to 0)
	candidateDeclareTime time.Time // the time when this will declare itself a candidate.
	//snapshotInProgress     []byte    // A snapshot that's being received through a sequence of InstallSnapshot RPCs.

	// VOLATILE ON LEADERS: reinitialized after election, nil on non-leaders
	nextIndex []int // for each server, index of the next Log entry to send to that server
	// nextIndex[me] doesn't matter. (initialized to leader last Log index + 1)
	matchIndex []int // for each server, index of highest Log entry known to be replicated on server
	// (initialized to 0, increases monotonically)
}
