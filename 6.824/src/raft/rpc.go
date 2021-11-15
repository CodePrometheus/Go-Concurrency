package raft

import "fmt"

type StateType uint64

const NULL int = -1

const (
	StateFollower StateType = iota
	StateCandidate
	StateLeader
)

// ApplyMsg
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
	CommandTerm  int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

func (msg ApplyMsg) String() string {
	if msg.CommandValid {
		return fmt.Sprintf("{Command:%v,CommandTerm:%v,CommandIndex:%v}", msg.Command, msg.CommandTerm, msg.CommandIndex)
	} else if msg.SnapshotValid {
		return fmt.Sprintf("{Snapshot:%v,SnapshotTerm:%v,SnapshotIndex:%v}", msg.Snapshot, msg.SnapshotTerm, msg.SnapshotIndex)
	} else {
		panic(fmt.Sprintf("unexpected ApplyMsg{CommandValid:%v,CommandTerm:%v,CommandIndex:%v,SnapshotValid:%v,SnapshotTerm:%v,SnapshotIndex:%v}", msg.CommandValid, msg.CommandTerm, msg.CommandIndex, msg.SnapshotValid, msg.SnapshotTerm, msg.SnapshotIndex))
	}
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	LeaderCommit int
	Entries      []Entry
}

func (args AppendEntriesArgs) String() string {
	return fmt.Sprintf("{Term:%v,LeaderId:%v,PrevLogIndex:%v,PrevLogTerm:%v,LeaderCommit:%v,Entries:%v}",
		args.Term, args.LeaderId, args.PrevLogIndex, args.PrevLogTerm, args.LeaderCommit, args.Entries)
}

type AppendEntriesReply struct {
	Term          int
	Success       bool
	ConflictIndex int
	ConflictTerm  int
}

func (reply AppendEntriesReply) String() string {
	return fmt.Sprintf("{Term:%v,Success:%v,ConflictIndex:%v,ConflictTerm:%v}",
		reply.Term, reply.Success, reply.ConflictIndex, reply.ConflictTerm)
}

type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

func (args RequestVoteArgs) String() string {
	return fmt.Sprintf("{Term:%v,CandidateId:%v,LastLogIndex:%v,LastLogTerm:%v}",
		args.Term, args.CandidateId, args.LastLogIndex, args.LastLogTerm)
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func (reply RequestVoteReply) String() string {
	return fmt.Sprintf("{Term:%v,VoteGranted:%v}", reply.Term, reply.VoteGranted)
}

type Entry struct {
	Index   int
	Term    int
	Command interface{}
}

func (entry Entry) String() string {
	return fmt.Sprintf("{Index:%v,Term:%v}", entry.Index, entry.Term)
}

type InstallSnapshotArgs struct {
	Term              int
	LeaderId          int
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte
}

func (args *InstallSnapshotArgs) String() string {
	return fmt.Sprintf("{Term:%v,LeaderId:%v,LastIncludedIndex:%v,LastIncludedTerm:%v,DataSize:%v}",
		args.Term, args.LeaderId, args.LastIncludedIndex, args.LastIncludedTerm, len(args.Data))
}

type InstallSnapShotReply struct {
	Term int
}

func (reply *InstallSnapShotReply) String() string {
	return fmt.Sprintf("{Term:%v}", reply.Term)
}
