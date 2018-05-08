package raft

import (
	"ad"
	"fmt"
)

// RequestVote RPC arguments structure.
type RequestVoteArgs struct {
	Term         int // candidate's term
	CandidateId  int
	LastLogIndex int // index of candidate’s last Log entry
	LastLogTerm  int // term of candidate’s last Log entry
}

// RequestVote RPC reply structure.
type RequestVoteReply struct {
	Term        int  // receiver's CurrentTerm, for candidate to update itself
	VoteGranted bool // true iff candidate received vote
	VoterId     int  // who voted
}

// handles making args for, sending, and receiving a requestVote RPC.
func (rf *Raft) sendRequestVote(peerNum int, repliesChan chan *RequestVoteReply) {
	rf.lock()

	assert(rf.me != peerNum) // voting for self is handled separately

	if !rf.isAlive {
		ad.DebugObj(rf, ad.TRACE, "Skipping requestVote to %d because I am dead", rf.me)
		rf.unlock()
		return
	}

	args := &RequestVoteArgs{}
	args.Term = rf.CurrentTerm
	args.CandidateId = rf.me
	args.LastLogIndex = rf.lastLogIndex()
	args.LastLogTerm = rf.Log.lastTerm()
	reply := &RequestVoteReply{}
	ad.DebugObj(rf, ad.RPC, "sending RequestVote to server %v", peerNum)
	rf.unlock()

	ok := rf.peers[peerNum].Call("Raft.RequestVote", args, reply)

	rf.lock()
	defer rf.unlock()

	rf.updateTermIfNecessary(reply.Term)
	if rf.CurrentElectionState != Candidate || rf.CurrentTerm != args.Term {
		ad.DebugObj(rf, ad.TRACE, "Election for term %d is over, abandoning response from server %v", args.Term, peerNum)
	} else if ok {
		// It's okay that unlocking comes after the channel push because the channel is guaranteed to never block
		// (because it has space for a reply from every peer)
		repliesChan <- reply
	} else {
		ad.DebugObj(rf, ad.RPC, "RequestVote from server %v did not succeed!", peerNum)
	}
}

// RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.lock()
	defer rf.writePersist()
	defer rf.unlock()

	if !rf.isAlive {
		ad.DebugObj(rf, ad.TRACE, "Ignoring RequestVote from %d because I am dead", args.CandidateId)
		return
	}

	ad.DebugObj(rf, ad.RPC, "Received RequestVote from %v in term %d", args.CandidateId, args.Term)

	reply.Term = rf.CurrentTerm
	reply.VoterId = rf.me
	var reason string

	receiverLastLogTerm := rf.Log.lastTerm()
	receiverLastLogIndex := rf.lastLogIndex()

	rf.updateTermIfNecessary(args.Term)

	switch {
	// maybe add this?
	// if i've heard from the leader less than (minimum election timeout) ago, automatically say no because
	// the candidate can't be connected to the same leader
	case args.Term < rf.CurrentTerm:
		reason = fmt.Sprintf("candidate's term %v < voter's term %v", args.Term, rf.CurrentTerm)
		reply.VoteGranted = false

	case args.Term == rf.CurrentTerm && rf.VotedFor != args.CandidateId && rf.VotedFor != -1:
		reason = fmt.Sprintf("voter already voted for someone else (%v) in term %d", rf.VotedFor, rf.CurrentTerm)
		reply.VoteGranted = false

	case receiverLastLogTerm > args.LastLogTerm:
		reason = fmt.Sprintf("voter's last Log entry has later term (%v) than candidate's (%v)",
			receiverLastLogTerm, args.LastLogTerm)
		reply.VoteGranted = false

	case receiverLastLogTerm == args.LastLogTerm && receiverLastLogIndex > args.LastLogIndex:
		reason = fmt.Sprintf("voter's Log (len %v) is longer than candidate's (len %v)", receiverLastLogIndex, args.LastLogIndex)
		reply.VoteGranted = false

	default:
		reply.VoteGranted = true
	}

	if reply.VoteGranted {
		rf.VotedFor = args.CandidateId
		rf.resetElectionTimeout()
		ad.DebugObj(rf, ad.RPC, "voting for %v and returning from RequestVote RPC", args.CandidateId)
		rf.writePersist()
	} else {
		ad.DebugObj(rf, ad.RPC, "not voting for %v in term %v because %v", args.CandidateId, args.Term, reason)
	}

}
