package raft

type StateType uint64

const NULL int = -1

const (
	StateFollower StateType = iota
	StateCandidate
	StateLeader
)

type AppendEntriesArgs struct {
	Term     int
	LeaderId int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type RequestVoteArgs struct {
	Term        int
	CandidateId int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type Entry struct {
	Index   int
	Term    int
	Command interface{}
}
