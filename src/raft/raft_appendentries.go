package raft

import (
	"ad"
	"fmt"
	"time"
)

type AppendEntriesArgs struct {
	Term         int        // leader's term
	LeaderID     int        // so follower can redirect clients
	PrevLogIndex int        // index of Log entry immediately preceding new ones
	PrevLogTerm  int        // term of previous Log entry
	Entries      []LogEntry // Log Entries to store (empty for heartbeat)
	LeaderCommit int        // leader's CommitIndex
}

type AppendEntriesReply struct {
	Term                        int  // follower's term
	Success                     bool // true iff follower's Log contained entry matching PrevLogIndex and PrevLogTerm
	ConflictingTerm             int  // term of the conflicting entry or -1 if there is no such entry
	FirstIndexOfConflictingTerm int  // index of the first entry of conflicting term or -1 if there is no such entry
	DesiredNextIndex            int  // Request to send entries starting with this index.
	DesiredNextIndexIsSet       bool // if false (by default), ignore DesiredNextIndex.
}

// constructs an AppendMessages and sends it to peerNum.
// set includeEntries=false to include no Log Entries.
func (rf *Raft) sendAppendEntries(peerNum int, includeEntries bool) {
	rf.lock()

	if rf.me == peerNum {
		// don't send an RPC to yourself, silly
		rf.unlock()
		return
	}
	if !rf.isAlive {
		ad.DebugObj(rf, ad.TRACE, "Skipping AppendEntries to %d because I am dead", peerNum, rf.me)
		rf.unlock()
		return
	}
	if rf.CurrentElectionState != Leader {
		ad.DebugObj(rf, ad.TRACE, "Server %d would send out a heartbeat to %d even though %d is not the leader! Instead, skipping heartbeat",
			rf.me, peerNum, rf.me)
		rf.unlock()
		return
	}
	args := &AppendEntriesArgs{}
	args.Term = rf.CurrentTerm
	args.LeaderID = rf.me
	args.LeaderCommit = rf.commitIndex

	args.PrevLogIndex = max(rf.nextIndex[peerNum]-1, 0)
	if args.PrevLogIndex < rf.lastIndexInSnapshot() {
		// send an InstallSnapshot instead
		ad.DebugObj(rf, ad.TRACE, "Would send an AppendEntries to %d with PrevLogIndex=%d, but already snapshotted "+
			"indices <= %d, sending InstallSnapshot instead", peerNum, args.PrevLogIndex, rf.lastIndexInSnapshot())
		go rf.sendInstallSnapshot(peerNum)
		rf.unlock()
		return
	}
	if rf.Log.indexIsUncompressed(args.PrevLogIndex) {
		args.PrevLogTerm = rf.Log.get(args.PrevLogIndex).Term
	} else {
		assertEquals(rf.Log.lastCompressedIndex(), args.PrevLogIndex)
		args.PrevLogTerm = rf.Log.lastCompressedTerm()
	}

	var entries []LogEntry
	if includeEntries {
		indexOfFirstEntry := args.PrevLogIndex + 1
		numMissingEntries := rf.Log.length() - indexOfFirstEntry + 1
		entries = make([]LogEntry, numMissingEntries)
		numCopiedEntries := copy(entries, rf.Log.getIndicesIncludingAndAfter(indexOfFirstEntry))
		assertEquals(numMissingEntries, numCopiedEntries)
	} else {
		// no need to do anything because an uninitialized slice is empty and ready to use
	}
	args.Entries = entries
	reply := &AppendEntriesReply{}

	sendTime := time.Now().Format("03:04:05.000")
	rfLastLogIndexBeforeSendingRPC := rf.lastLogIndex()
	ad.DebugObj(rf, ad.RPC, "Sending AppendEntries with %d entries to %v", len(args.Entries), peerNum)
	rf.unlock()

	ok := rf.peers[peerNum].Call("Raft.AppendEntries", args, reply)

	rf.lock()
	defer rf.unlock()

	rf.updateTermIfNecessary(reply.Term)

	if !rf.isAlive {
		// don't even print a trace because you're DEAD
		return
	}
	if !(rf.CurrentElectionState == Leader) {
		ad.DebugObj(rf, ad.TRACE, "Ignoring AppendEntries reply from %d that I sent in term %d because I am no longer the leader",
			peerNum, args.Term)
		return
	}
	if rf.CurrentTerm > reply.Term {
		ad.DebugObj(rf, ad.TRACE, "Ignoring AppendEntries reply from %d (term %d) because I am in greater term %d",
			peerNum, reply.Term, rf.CurrentTerm)
		return
	}

	ad.DebugObj(rf, ad.TRACE, "received %+v, ok=%t from AppendEntries to %d sent in term %d at %v",
		reply, ok, peerNum, args.Term, sendTime)
	if ok && reply.Success {
		rf.nextIndex[peerNum] = rfLastLogIndexBeforeSendingRPC + 1
		rf.matchIndex[peerNum] = rfLastLogIndexBeforeSendingRPC
		ad.DebugObj(rf, ad.TRACE, "reply success, nextIndex=%+v, matchIndex=%+v", rf.nextIndex, rf.matchIndex)

		// If there exists an N such that N > commitIndex, a majority of matchIndex[i] ≥ N, and
		// Log[N].term == CurrentTerm: set commitIndex = N (§5.3, §5.4).
		for n := rf.commitIndex + 1; n <= rf.lastLogIndex(); n++ {
			// if a majority of matchIndex[i] >= N and Log[N].term == CurrentTerm
			ad.DebugObj(rf, ad.TRACE, "Considering updating rf.commitIndex from %d to %d...", rf.commitIndex, n)
			if rf.Log.get(n).Term == rf.CurrentTerm {
				numMatchIndexAtLeastN := 0
				ad.DebugObj(rf, ad.TRACE, "matchIndex=%+v", rf.matchIndex)
				for peerNum, _ := range rf.peers {
					if rf.matchIndex[peerNum] >= n {
						numMatchIndexAtLeastN++
					}
				}
				if numMatchIndexAtLeastN >= rf.majoritySize() {
					ad.DebugObj(rf, ad.TRACE, "%d peers matchIndex %d, increasing commitIndex from %d to %d",
						numMatchIndexAtLeastN, n, rf.commitIndex, n)
					rf.commitIndex = n
					go func() { rf.toApply <- true }()
				} else {
					ad.DebugObj(rf, ad.TRACE, "only %d peers match up to index %d, can't commit", numMatchIndexAtLeastN, n)
					break
				}
			} else {
				ad.DebugObj(rf, ad.TRACE, "not choosing because Log[%d] has wrong term %d (instead of %d)",
					n, rf.Log.get(n).Term, rf.CurrentTerm)
			}
		}
	} else {
		// If the follower told us exactly what they want us to send
		if reply.DesiredNextIndexIsSet {
			assert(reply.DesiredNextIndex > 0)
			// it could be equal to logLength + 1 if they already have all our entries in a snapshot
			assert(reply.DesiredNextIndex <= rf.Log.length()+1)
			ad.DebugObj(rf, ad.TRACE, "At follower's request, setting nextIndex=%d", reply.DesiredNextIndex)
			rf.nextIndex[peerNum] = reply.DesiredNextIndex
		} else {
			// we need to figure out what to send for ourselves
			leaderHasEntriesWithConflictingTerm := false
			leaderLastIndexWithConflictingTerm := -1

			for i := rf.lastLogIndex(); i > rf.Log.lastCompressedIndex(); i-- {
				if rf.Log.get(i).Term == reply.ConflictingTerm {
					leaderHasEntriesWithConflictingTerm = true
					leaderLastIndexWithConflictingTerm = i
				}
			}

			if leaderHasEntriesWithConflictingTerm {
				rf.nextIndex[peerNum] = leaderLastIndexWithConflictingTerm
				ad.DebugObj(rf, ad.TRACE, "leader has Entries with conflicting term %d, setting nextIndex[%d] to %d",
					reply.ConflictingTerm, peerNum, rf.nextIndex[peerNum])
			} else {
				rf.nextIndex[peerNum] = reply.FirstIndexOfConflictingTerm
				ad.DebugObj(rf, ad.TRACE, "leader does not have Entries with conflicting term %d, setting nextIndex[%d] to %d",
					reply.ConflictingTerm, peerNum, rf.nextIndex[peerNum])
			}

			rf.nextIndex[peerNum] = max(1, rf.nextIndex[peerNum]) // just in case
		}
		ad.DebugObj(rf, ad.TRACE, "AppendEntries to %d failed, setting nextIndex[%d] to %d and trying again.", peerNum, peerNum, rf.nextIndex[peerNum])
		go rf.sendAppendEntries(peerNum, includeEntries)
	}
}

// AppendEntries RPC handler.
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.lock()
	defer rf.unlock()

	if !rf.isAlive {
		ad.DebugObj(rf, ad.TRACE, "Ignoring AppendEntries from %d because I am dead", args.LeaderID)
		return
	}

	debugStr := fmt.Sprintf("AppendEntries from %d", args.LeaderID)
	ad.DebugObj(rf, ad.RPC, "received %v%+v", debugStr, args)

	rf.updateTermIfNecessary(args.Term)
	reply.Term = rf.CurrentTerm

	var reason string
	switch {
	case args.Term < rf.CurrentTerm: // implementation step 1
		reason = fmt.Sprintf("I have greater term %d", rf.CurrentTerm)
		reply.Success = false

	case rf.lastLogIndex() < args.PrevLogIndex: // implementation step 2
		reason = fmt.Sprintf("Log only has indices up to %d", rf.lastLogIndex())
		reply.Success = false

	case args.PrevLogIndex < rf.lastIndexInSnapshot():
		reply.DesiredNextIndex = rf.lastIndexInSnapshot() + 1
		reply.DesiredNextIndexIsSet = true
		reason = fmt.Sprintf("prevLogIndex is already compressed into a snapshot, "+
			"setting DesiredNextIndex to lastIndexInSnapshot+1=%d", reply.DesiredNextIndex)
		reply.Success = false

	case args.PrevLogIndex == rf.lastIndexInSnapshot() &&
		rf.lastIndexInSnapshot() < rf.commitIndex:
		reply.DesiredNextIndex = rf.commitIndex + 1
		reply.DesiredNextIndexIsSet = true
		reason = fmt.Sprintf("prevLogIndex = lastIndexInSnapshot, but entries after that have already been committed, "+
			"setting desiredNextIndex to commitIndex+1=%d", reply.DesiredNextIndex)
		reply.Success = false

	case (args.PrevLogIndex >= 0) &&
		(args.PrevLogIndex <= rf.lastLogIndex()) &&
		rf.Log.indexIsUncompressed(args.PrevLogIndex) &&
		rf.Log.get(args.PrevLogIndex).Term != args.PrevLogTerm:
		reason = fmt.Sprintf("Log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm")
		reply.Success = false

	case args.Term == rf.CurrentTerm && rf.CurrentElectionState == Leader:
		panic("received AppendEntries from a different server but in the same term!")

	default:
		reply.Success = true
	}

	if reply.Success {
		assert(rf.CurrentElectionState != Leader)
		rf.CurrentElectionState = Follower

		// Log optimization: these should default to -1, not 0 (the default for integers)
		reply.ConflictingTerm = -1
		reply.FirstIndexOfConflictingTerm = -1

		assert(args.PrevLogIndex >= -1)
		ad.DebugObj(rf, ad.TRACE, "log=%+v", rf.Log)
		for newEntriesIndex, newEntry := range args.Entries {
			indexInLog := args.PrevLogIndex + newEntriesIndex + 1
			assertEquals(indexInLog, newEntry.Index)
			if indexInLog <= rf.lastLogIndex() {
				// This is safe because if indexInLog was already compressed, the RPC would have been rejected above.
				entry := rf.Log.get(indexInLog)
				if (entry.Index == newEntry.Index) && (entry.Term != newEntry.Term) {
					ad.DebugObj(rf, ad.RPC, "found conflicting entry %+v, differs from new entry %+v, deleting it and all later Entries (%d)",
						entry, newEntry, len(rf.Log.getIndicesIncludingAndAfter(entry.Index)))

					// Log optimization
					reply.ConflictingTerm = entry.Term
					i := args.PrevLogIndex
					for (i > 0) && (rf.Log.indexIsUncompressed(i)) && (rf.Log.get(i).Term == reply.ConflictingTerm) {
						ad.DebugObj(rf, ad.TRACE, "Log.get(%d)=%+v, in conflicting term %d", i, rf.Log.get(i), reply.ConflictingTerm)
						i--
					}
					reply.FirstIndexOfConflictingTerm = i

					rf.Log.truncateAfter(entry.Index - 1)
					ad.DebugObj(rf, ad.TRACE, "log=%+v", rf.Log)
					ad.DebugObj(rf, ad.TRACE, "ConflictingTerm=%d and firstIndexOfConflictingTerm=%d",
						reply.ConflictingTerm, reply.FirstIndexOfConflictingTerm)

					rf.writePersist()
					reply.Success = false
				}
			} // end if
		} // end for
	} else {
		ad.DebugObj(rf, ad.RPC, "Rejecting AppendEntries because %v", reason)
		return
	}

	// step 4
	// only append Entries that are not already in the log
	if len(args.Entries) > 0 {
		var toAppend []LogEntry
		ad.DebugObj(rf, ad.TRACE, "Beginning to add %+v to my Log", args.Entries)
		for _, entryFromLeader := range args.Entries {
			entryAlreadyInLog := false
			// We know that these can't be compressed entries; if they were, then args.PrevLogIndex would be
			// < rf.lastIndexInSnapshot() and the RPC would have been rejected.
			for _, uncompressedEntryInLog := range rf.Log.UncompressedEntries {
				//if entryFromLeader == uncompressedEntryInLog {
				if LogEntryEquals(entryFromLeader, uncompressedEntryInLog) {
					ad.DebugObj(rf, ad.TRACE, "%+v is already in the log (uncompressed)", entryFromLeader)
					entryAlreadyInLog = true
				}
			}
			if !entryAlreadyInLog {
				toAppend = append(toAppend, entryFromLeader)
				ad.DebugObj(rf, ad.TRACE, "appending %+v", entryFromLeader)
			} else {
				ad.DebugObj(rf, ad.TRACE, "not appending %+v because i already had it", entryFromLeader)
			}
		}
		rf.Log.appendAll(toAppend)
		rf.writePersist()
		ad.DebugObj(rf, ad.TRACE, "done appending, Log=%+v", rf.Log)
	} else {
		ad.DebugObj(rf, ad.TRACE, "No Log Entries in AppendEntries")
	}

	// step 5
	if args.LeaderCommit > rf.commitIndex {
		newCommitIndex := min(args.LeaderCommit, rf.lastLogIndex())
		ad.DebugObj(rf, ad.TRACE, "Updating commitIndex from %d to %d", rf.commitIndex, newCommitIndex)
		assert(newCommitIndex >= rf.commitIndex) // make sure commitIndex only increases
		rf.commitIndex = newCommitIndex
		if rf.commitIndex > rf.lastApplied {
			go func() { rf.toApply <- true }()
		}
	}

	rf.resetElectionTimeout()

	ad.DebugObj(rf, ad.RPC, "done processing %v, postponing running election until %v",
		debugStr, rf.candidateDeclareTime.Format("05.000"))
}
