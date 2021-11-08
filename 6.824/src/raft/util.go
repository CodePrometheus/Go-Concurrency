package raft

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const Debug = true

const (
	HeartbeatTimeout = 125
	ElectionTimeout  = 1000
)

func DPrintf(format string, a ...interface{}) {
	if Debug {
		log.Printf(format, a...)
	}
}

type lockedRand struct {
	mu   sync.Mutex
	rand *rand.Rand
}

func (r *lockedRand) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Intn(n)
}

var globalRand = &lockedRand{
	rand: rand.New(rand.NewSource(time.Now().UnixNano())),
}

func RandomElectionTimeout() time.Duration {
	return time.Duration(ElectionTimeout+globalRand.Intn(ElectionTimeout)) * time.Millisecond
}

func SingleHeartbeatTimeout() time.Duration {
	return time.Duration(HeartbeatTimeout) * time.Millisecond
}

// toString()
func (state StateType) String() string {
	switch state {
	case StateFollower:
		return "Follower"
	case StateCandidate:
		return "Candidate"
	case StateLeader:
		return "Leader"
	}
	panic(fmt.Sprintf("Unexpected NodeState %d", state))
}

func (entry Entry) String() string {
	return fmt.Sprintf("{Index: %v, Term:%v}", entry.Index, entry.Term)
}

func (reply RequestVoteReply) String() string {
	return fmt.Sprintf("{Term: %v, VoteGranted: %v}", reply.Term, reply.VoteGranted)
}

func (args RequestVoteArgs) String() string {
	return fmt.Sprintf("{Term: %d, CandidateId: %d}", args.Term, args.CandidateId)
}
