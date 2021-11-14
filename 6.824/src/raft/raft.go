package raft

// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.

import (
	"6.824/labgob"
	"bytes"
	//	"bytes"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

// ApplyMsg
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// Raft A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.RWMutex        // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	applyCh   chan ApplyMsg // 当Leader apply一个log entry到state machine以后，会通知该channel；这样，client只需要通过监控applyCh是否有更新即可知道是否command已commit成功
	applyCond *sync.Cond

	state       StateType
	currentTerm int
	votedFor    int

	commitIndex int
	lastApplied int

	log        []Entry
	nextIndex  []int
	matchIndex []int

	heartbeatTimeout *time.Timer
	electionTimeout  *time.Timer
}

// GetState return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.RLock()
	defer rf.mu.RUnlock()
	return rf.currentTerm, rf.state == StateLeader
}

func (rf *Raft) scheduleState(state StateType) {
	if state == rf.state {
		return
	}
	DPrintf("{ Node %d } change state from %v to %v in Term %d", rf.me, rf.state, state, rf.currentTerm)
	rf.state = state
	switch state {
	// 转换后调正对应角色heartbeat和electionTime操作
	case StateFollower:
		rf.heartbeatTimeout.Stop()
		rf.electionTimeout.Reset(RandomElectionTimeout())
	case StateLeader:
		lastLog := rf.getLastLog()
		for i := 0; i < len(rf.peers); i++ {
			rf.nextIndex[i], rf.matchIndex[i] = lastLog.Index+1, 0
		}
		rf.electionTimeout.Stop()
		rf.heartbeatTimeout.Reset(SingleHeartbeatTimeout())
	}
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 {
		return
	}
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	var currentTerm, voteFor int
	var log []Entry
	if d.Decode(&currentTerm) != nil ||
		d.Decode(&voteFor) != nil ||
		d.Decode(&log) != nil {
		DPrintf("{ Node %v } restores persisted state failed", rf.me)
	}
	rf.currentTerm, rf.votedFor, rf.log = currentTerm, voteFor, log
	// 日志末尾
	rf.lastApplied, rf.commitIndex = rf.log[0].Index, rf.log[0].Index
}

// CondInstallSnapshot A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// Snapshot
// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

// RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()
	defer DPrintf("{ Node %v }'s state is {state %v,term %v,commitIndex %v,lastApplied %v,firstLog %v,"+
		"lastLog %v} before processing requestVoteRequest %v and reply requestVoteResponse %v",
		rf.me, rf.state, rf.currentTerm, rf.commitIndex, rf.lastApplied, rf.getFirstLog(), rf.getLastLog(),
		args, reply)

	// 满足以下函数直接结束
	if args.Term < rf.currentTerm || (args.Term == rf.currentTerm &&
		rf.votedFor != NULL &&
		rf.votedFor != args.CandidateId) {
		reply.Term, reply.VoteGranted = rf.currentTerm, false
		return
	}

	if args.Term > rf.currentTerm {
		rf.scheduleState(StateFollower)
		rf.currentTerm, rf.votedFor = args.Term, NULL
	}

	if !rf.isLogUpToDate(args.LastLogTerm, args.LastLogIndex) {
		reply.Term, reply.VoteGranted = rf.currentTerm, false
		return
	}

	// ok
	rf.votedFor = args.CandidateId
	rf.electionTimeout.Reset(RandomElectionTimeout())
	reply.Term, reply.VoteGranted = rf.currentTerm, true
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in 6.824/labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	return rf.peers[server].Call("Raft.RequestVote", args, reply)
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	return rf.peers[server].Call("Raft.AppendEntries", args, reply)
}

func (rf *Raft) StartElection() {
	arg := rf.sendVoteRequest()
	DPrintf("{ Node %v } starts election with RequestVoteArgs %v", rf.me, arg)
	// closure闭包
	grantedVotes := 1
	rf.votedFor = rf.me
	rf.persist()
	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		// 并行异步投票
		go func(peer int) {
			reply := new(RequestVoteReply)
			if rf.sendRequestVote(peer, arg, reply) {
				rf.mu.Lock()
				defer rf.mu.Unlock()
				DPrintf("{ Node %v } receives RequestVoteReply %v from { Node %v }"+
					" after sending RequestVoteArgs %v in term %v", rf.me, reply,
					peer, arg, rf.currentTerm)
				if rf.currentTerm == arg.Term && rf.state == StateCandidate {
					if reply.VoteGranted {
						grantedVotes += 1
						if grantedVotes > len(rf.peers)/2 {
							DPrintf("{ Node %v } receives majority votes in term %v",
								rf.me, rf.currentTerm)
							rf.scheduleState(StateLeader)
							rf.BroadcastHeartbeat(true)
						}
					} else if reply.Term > rf.currentTerm {
						DPrintf("{ Node %v } finds a new leader { Node %v } with term %v and "+
							"steps down in term %v", rf.me, peer, reply.Term, rf.currentTerm)
						rf.scheduleState(StateFollower)
						rf.currentTerm, rf.votedFor = reply.Term, NULL
						rf.persist()
					}
				}
			}
		}(peer)
	}
}

func (rf *Raft) sendVoteRequest() *RequestVoteArgs {
	return &RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: rf.getLastLog().Index,
		LastLogTerm:  rf.getLastLog().Term,
	}
}

// Start
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at if it's ever committed.
// the second return value is the current term.
// the third return value is true if this server believes it is he leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.state != StateLeader {
		return NULL, NULL, false
	}
	newLog := rf.appendNewLog(command)
	rf.BroadcastHeartbeat(false)
	return newLog.Index, newLog.Term, true
}

func (rf *Raft) appendNewLog(command interface{}) Entry {
	lastLog := rf.getLastLog()
	newLog := Entry{
		Index:   lastLog.Index + 1,
		Term:    rf.currentTerm,
		Command: command,
	}
	rf.log = append(rf.log, newLog)
	rf.matchIndex[rf.me] = newLog.Index
	rf.nextIndex[rf.me] = newLog.Index + 1
	rf.persist()
	return newLog
}

// Kill
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	return atomic.LoadInt32(&rf.dead) == 1
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
// code here to check if a leader election should
// be started and to randomize sleeping time using time.Sleep().
func (rf *Raft) ticker() {
	for !rf.killed() {
		// two case
		select {
		case <-rf.electionTimeout.C:
			rf.mu.Lock()
			rf.scheduleState(StateCandidate)
			rf.currentTerm += 1
			rf.StartElection()
			rf.electionTimeout.Reset(RandomElectionTimeout())
			rf.mu.Unlock()
		case <-rf.heartbeatTimeout.C:
			rf.mu.Lock()
			if rf.state == StateLeader {
				rf.BroadcastHeartbeat(true)
				rf.heartbeatTimeout.Reset(SingleHeartbeatTimeout())
			}
			rf.mu.Unlock()
		}
	}
}

func (rf *Raft) BroadcastHeartbeat(heartBeat bool) {
	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		if heartBeat {
			go rf.handleHeartbeat(peer)
		}
	}
}

func (rf *Raft) handleHeartbeat(peer int) {
	rf.mu.RLock()
	if rf.state != StateLeader {
		rf.mu.RUnlock()
		return
	}

	prevLogIndex := rf.nextIndex[peer] - 1
	if prevLogIndex < rf.getFirstLog().Index {
	} else {
		arg := rf.sendAppendEntriesRequest(prevLogIndex)
		rf.mu.RUnlock()
		reply := new(AppendEntriesReply)
		if rf.sendAppendEntries(peer, arg, reply) {
			rf.mu.Lock()
			rf.handleAppendEntriesReply(peer, arg, reply)
			rf.mu.Unlock()
		}
	}
}

func (rf *Raft) sendAppendEntriesRequest(preLogIdx int) *AppendEntriesArgs {
	firstIdx := rf.getFirstLog().Index
	entries := make([]Entry, len(rf.log[preLogIdx+1-firstIdx:]))
	copy(entries, rf.log[preLogIdx+1-firstIdx:])
	return &AppendEntriesArgs{
		Term:         rf.currentTerm,
		LeaderId:     rf.me,
		PrevLogIndex: preLogIdx,
		PrevLogTerm:  rf.log[preLogIdx-firstIdx].Term,
		LeaderCommit: rf.commitIndex,
		Entries:      entries,
	}
}

func (rf *Raft) handleAppendEntriesReply(peer int, arg *AppendEntriesArgs, reply *AppendEntriesReply) {
	if rf.state == StateLeader && rf.currentTerm == arg.Term {
		if reply.Success {
			rf.matchIndex[peer] = arg.PrevLogIndex + len(arg.Entries)
			rf.nextIndex[peer] = rf.matchIndex[peer] + 1
			rf.advanceCommitIndexForLeader()
		} else {
			if reply.Term > rf.currentTerm {
				rf.scheduleState(StateFollower)
				rf.currentTerm, rf.votedFor = reply.Term, NULL
				rf.persist()
			} else if reply.Term == rf.currentTerm {
				rf.nextIndex[peer] = reply.ConflictIndex
				if reply.ConflictTerm != -1 {
					firstIdx := rf.getFirstLog().Index
					for i := arg.PrevLogIndex; i >= firstIdx; i-- {
						if rf.log[i-firstIdx].Term == reply.ConflictTerm {
							rf.nextIndex[peer] = i + 1
							break
						}
					}
				}
			}
		}
	}
	DPrintf("{ Node %v }'s state is {state %v,term %v,commitIndex %v,lastApplied %v,firstLog %v,"+
		"lastLog %v} after handling AppendEntriesReply %v for AppendEntriesArgs %v",
		rf.me, rf.state, rf.currentTerm, rf.commitIndex, rf.lastApplied, rf.getFirstLog(), rf.getLastLog(),
		reply, arg)
}

func (rf *Raft) advanceCommitIndexForLeader() {
	n := len(rf.matchIndex)
	res := make([]int, n)
	copy(res, rf.matchIndex)
	insertionSort(res)
	medianIndex := res[n-(n/2+1)]
	if medianIndex > rf.commitIndex {
		if rf.matchLog(rf.currentTerm, medianIndex) {
			DPrintf("{ Node %d } advance commitIndex from %d to %d with matchIndex %v in"+
				" term %d", rf.me, rf.commitIndex, medianIndex, rf.matchIndex, rf.currentTerm)
			rf.commitIndex = medianIndex
			rf.applyCond.Signal()
		} else {
			DPrintf("{ Node %d } can not advance commitIndex from %d because the term "+
				"of medianIndex %d is not equal to currentTerm %d", rf.me, rf.commitIndex,
				medianIndex, rf.currentTerm)
		}
	}
}

// AppendEntries 投票 & 心跳机制
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()
	defer DPrintf("{ Node %v }'s state is {state %v,term %v,commitIndex %v,lastApplied %v,firstLog %v,lastLog %v}"+
		" before processing AppendEntriesArgs %v and reply AppendEntriesReply %v", rf.me, rf.state, rf.currentTerm,
		rf.commitIndex, rf.lastApplied, rf.getFirstLog(), rf.getLastLog(), args, reply)

	// 拒绝
	if args.Term < rf.currentTerm {
		reply.Term, reply.Success = rf.currentTerm, false
		return
	}

	if args.Term > rf.currentTerm {
		rf.currentTerm, rf.votedFor = args.Term, NULL
	}

	rf.scheduleState(StateFollower)
	rf.electionTimeout.Reset(RandomElectionTimeout())

	if args.PrevLogIndex < rf.getFirstLog().Index {
		reply.Term, reply.Success = 0, false
		DPrintf("{ Node %v } receives unexpected AppendEntriesArgs %v "+
			"from { Node %v } because prevLogIndex %v < firstLogIndex %v", rf.me,
			args, args.LeaderId, args.PrevLogIndex, rf.getFirstLog().Index)
		return
	}

	if !rf.matchLog(args.PrevLogTerm, args.PrevLogIndex) {
		reply.Term, reply.Success = rf.currentTerm, false
		lastIndex := rf.getLastLog().Index
		if lastIndex < args.PrevLogIndex {
			reply.ConflictIndex, reply.ConflictTerm = lastIndex+1, -1
		} else {
			firstIdx := rf.getFirstLog().Index
			reply.ConflictTerm = rf.log[args.PrevLogIndex-firstIdx].Term
			conflictIdx := args.PrevLogIndex - 1
			for conflictIdx >= firstIdx &&
				rf.log[conflictIdx-firstIdx].Term == reply.ConflictTerm {
				conflictIdx--
			}
			reply.ConflictIndex = conflictIdx
		}
		return
	}

	firstIdx := rf.getFirstLog().Index
	for idx, entry := range args.Entries {
		if entry.Index-firstIdx >= len(rf.log) ||
			rf.log[entry.Index-firstIdx].Term != entry.Term {
			rf.log = curEntries(append(rf.log[:entry.Index-firstIdx],
				args.Entries[idx:]...))
			break
		}
	}

	rf.advanceCommitIndexForFollower(args.LeaderCommit)

	reply.Term, reply.Success = rf.currentTerm, true
}

func (rf *Raft) applier() {
	for !rf.killed() {
		rf.mu.Lock()
		for rf.lastApplied >= rf.commitIndex {
			rf.applyCond.Wait()
		}

		firstIndex, commitIndex, lastApplied := rf.getFirstLog().Index, rf.commitIndex, rf.lastApplied
		entries := make([]Entry, commitIndex-lastApplied)
		copy(entries, rf.log[lastApplied+1-firstIndex:commitIndex+1-firstIndex])
		rf.mu.Unlock()

		for _, entry := range entries {
			rf.applyCh <- ApplyMsg{
				CommandValid: true,
				Command:      entry.Command,
				CommandIndex: entry.Index,
			}
		}

		rf.mu.Lock()
		DPrintf("{ Node %v } applies entries %v-%v in term %v", rf.me, rf.lastApplied,
			commitIndex, rf.currentTerm)
		// lastApplied <= commitIndex
		rf.lastApplied = Max(rf.lastApplied, commitIndex)
		rf.mu.Unlock()
	}
}

func (rf *Raft) advanceCommitIndexForFollower(leaderCommit int) {
	// AppendEntries RPC rule 5
	newCommitIdx := Min(leaderCommit, rf.getLastLog().Index)
	if newCommitIdx > rf.commitIndex {
		DPrintf("{ Node %d } advance commitIndex from %d to %d with leaderCommit %d in term %d",
			rf.me, rf.commitIndex, newCommitIdx, leaderCommit, rf.currentTerm)
		rf.commitIndex = newCommitIdx
		rf.applyCond.Signal()
	}
}

func curEntries(entries []Entry) []Entry {
	const lenMultiple = 2
	if len(entries)*lenMultiple < cap(entries) {
		newEntries := make([]Entry, len(entries))
		copy(newEntries, entries)
		return newEntries
	}
	return entries
}

// Make the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{
		peers:            peers,
		persister:        persister,
		me:               me,
		dead:             0,
		state:            StateFollower,
		currentTerm:      0,
		votedFor:         NULL,
		heartbeatTimeout: time.NewTimer(SingleHeartbeatTimeout()),
		electionTimeout:  time.NewTimer(RandomElectionTimeout()),
		log:              make([]Entry, 1),
		nextIndex:        make([]int, len(peers)),
		matchIndex:       make([]int, len(peers)),
		applyCh:          applyCh,
	}

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	rf.applyCond = sync.NewCond(&rf.mu)

	// start ticker goroutine to start elections
	go rf.ticker()

	go rf.applier()

	return rf
}
