package raft

import (
	"labrpc"
	_ "net/http/pprof"
	"time"
)

// return CurrentTerm and whether this server believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.lock()
	defer rf.unlock()
	isLeader := rf.CurrentElectionState == Leader
	term := rf.CurrentTerm
	return term, isLeader
}

// Start agreement on a command to be appended to the Log.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.lock()
	defer rf.unlock()

	if rf.CurrentElectionState != Leader {
		debug(rf, TRACE, "Rejecting Start(%+v) because I am not the leader", command)
		return 0, 0, false
	}
	debug(rf, TRACE, "Received Start(%+v)", command)

	// +1 because it will go after the current last entry
	entry := LogEntry{rf.CurrentTerm, command, rf.Log.length() + 1}
	rf.Log.append(entry)
	if rf.CurrentElectionState == Leader {
		rf.matchIndex[rf.me] = rf.Log.length()
	}
	rf.writePersist()

	debug(rf, TRACE, "Sending new Log message to peers")
	for peerNum, _ := range rf.peers {
		go rf.sendAppendEntries(peerNum, true)
	}

	index := rf.lastLogIndex()
	term := rf.CurrentTerm
	isLeader := true

	debug(rf, RPC, "returning (%d, %d, %t) from Start(%+v)", index, term, isLeader, command)
	debug(rf, TRACE, "Log=%+v", rf.Log)

	return index, term, isLeader
}

// The tester calls Kill() when a Raft instance won't be needed again.
func (rf *Raft) Kill() {
	rf.lock()
	debug(rf, RPC, "Dying")
	rf.isAlive = false
	rf.unlock()
}

// Keeps track of applying Entries to the state machine and related paraphernalia.
func (rf *Raft) ApplierThread() {
	for {
		select {
		case <-rf.toApply:
			rf.lock()

			if !rf.isAlive {
				rf.unlock()
				return
			}

			if rf.commitIndex > rf.lastApplied {
				debug(rf, TRACE, "ApplierThread has awoken! Ready to apply indices up to %d", rf.commitIndex)

				for rf.lastApplied < rf.commitIndex {
					indexToApply := rf.lastApplied + 1
					entryToApply := rf.Log.get(indexToApply)
					applyMsg := ApplyMsg{true, entryToApply.Command, indexToApply, rf.CurrentTerm, COMMAND}
					debug(rf, TRACE, "About to apply %+v at index %d", entryToApply, indexToApply)
					rf.lastApplied = indexToApply
					rf.unlock()

					rf.applyCh <- applyMsg

					rf.lock() // for the next iteration
				}

				debug(rf, TRACE, "ApplierThread going back to sleep.")
			}

			rf.unlock()
		}
	}
}

// Keeps track of running for election every so often.
func (rf *Raft) ElectionThread() {
	for {
		rf.lock()

		if rf.CurrentElectionState == Leader {
			debug(rf, TRACE, "ElectionThread going back to sleep until not the leader.")
			rf.unlock()

			// blocking read
			<-rf.becomeFollower

			rf.lock()
			debug(rf, RPC, "Becoming Follower")
			rf.CurrentElectionState = Follower
			rf.writePersist()
			rf.resetElectionTimeout()
		}

		if !rf.isAlive {
			rf.unlock()
			return
		}

		if time.Now().After(rf.candidateDeclareTime) {
			debug(rf, TRACE, "I should run for election")
			go rf.runForElection()
			rf.resetElectionTimeout()
		}
		sleepDuration := time.Until(rf.candidateDeclareTime)
		debug(rf, TRACE, "ElectionThread going back to sleep for %v", sleepDuration.String())
		rf.unlock()
		time.Sleep(sleepDuration)
	}
}

// Keeps track of sending out heartbeats.
func (rf *Raft) HeartbeatThread() {
	for {
		rf.lock()
		if rf.CurrentElectionState != Leader {
			debug(rf, TRACE, "HeartbeatThread waiting until is leader")
			rf.unlock()

			// blocking read
		waitForBecomeLeader:
			term := <-rf.becomeLeader

			rf.lock()
			if term < rf.CurrentTerm {
				// I became leader in a previous term but then advanced to my current term before I
				// noticed I became a leader, so instead I should become a follower.
				debug(rf, WARN, "Just noticed that I won election in term %d, but it's now term %d, so I'll stay a follower",
					term, rf.CurrentTerm)
				assert(rf.CurrentElectionState != Leader)
				rf.unlock()
				goto waitForBecomeLeader
			}

			// term > rf.CurrentTerm wouldn't make any sense
			assert(term == rf.CurrentTerm)
			debug(rf, RPC, "Becoming leader")
			rf.CurrentElectionState = Leader
			rf.writePersist()
			for peerNum, _ := range rf.peers {
				rf.nextIndex[peerNum] = rf.lastLogIndex() + 1
				rf.matchIndex[peerNum] = 0
			}
			rf.matchIndex[rf.me] = rf.Log.length()
		}

		if !rf.isAlive {
			rf.unlock()
			return
		}

		debug(rf, RPC, "Sending heartbeats. commitIndex=%+v, nextIndex=%+v, matchIndex=%+v",
			rf.commitIndex, rf.nextIndex, rf.matchIndex)
		for peerNum, _ := range rf.peers {
			go rf.sendAppendEntries(peerNum, true)
		}
		rf.unlock()
		time.Sleep(getHeartbeatTimeout())

	}
}

// Blocks until the election is won or lost
func (rf *Raft) runForElection() {
	rf.lock()
	rf.CurrentTerm += 1
	rf.VotedFor = -1
	rf.CurrentElectionState = Candidate
	debug(rf, RPC, "Starting election and advancing term to %d", rf.CurrentTerm)
	rf.writePersist()
	repliesChan := make(chan *RequestVoteReply, len(rf.peers)-1)
	// The term the election was started in
	electionTerm := rf.CurrentTerm
	rf.unlock()

	for peerNum, _ := range rf.peers {
		if peerNum == rf.me {
			rf.lock()
			rf.VotedFor = rf.me
			debug(rf, TRACE, "voting for itself")
			rf.writePersist()
			rf.unlock()
		} else {
			go func(peerNum int, repliesChan chan *RequestVoteReply) {
				rf.sendRequestVote(peerNum, repliesChan)
			}(peerNum, repliesChan)
		}
	}

	yesVotes := 1 // from yourself
	noVotes := 0
	requiredToWin := rf.majoritySize()
	for range rf.peers {
		reply := <-repliesChan

		rf.lock()
		assert(rf.CurrentElectionState != Leader)
		if rf.CurrentTerm != electionTerm {
			debug(rf, TRACE, "advanced to term %d while counting results of election for term %d. "+
				"Abandoning election.")
			rf.unlock()
			return
		}

		if reply.VoteGranted {
			yesVotes++
		} else {
			noVotes++
		}

		debug(rf, TRACE, "Got %+v from server %d, yes votes now at %d out of a required %d",
			reply, reply.VoterId, yesVotes, requiredToWin)
		if yesVotes >= requiredToWin {
			debug(rf, RPC, "Won election!")
			// non-blocking send
			// send the term number to prevent a bug where the raft advances to a new term before it notices it's
			// become a leader, so it becomes a second false leader.
			go func(term int) { rf.becomeLeader <- term }(rf.CurrentTerm)
			rf.unlock()
			return
		} else if noVotes >= requiredToWin {
			debug(rf, RPC, "Got %d no votes, can't win election. Reverting to follower", noVotes)
			rf.CurrentElectionState = Follower
			rf.writePersist()
			rf.unlock()
			return
		} else {
			rf.unlock()
			// wait for more votes
		}
	}
}

// Create a raft server.

func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.lock() // i don't think this matters but i'm not taking chances

	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.isAlive = true
	rf.applyCh = applyCh

	rf.toApply = make(chan bool)
	rf.becomeLeader = make(chan int)
	rf.becomeFollower = make(chan bool)

	rf.VotedFor = -1
	rf.Log = makeEmptyLogOne()
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.CurrentElectionState = Follower
	rf.candidateDeclareTime = time.Now().Add(getElectionTimeout())
	rf.CurrentTerm = 0
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// store the state in case we crash immediately
	rf.writePersist()
	rf.unlock()

	go rf.ElectionThread()
	go rf.HeartbeatThread()
	go rf.ApplierThread()
	return rf
}
