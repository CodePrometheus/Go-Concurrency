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

func (rf *Raft) getLastLog() Entry {
	return rf.log[len(rf.log)-1]
}

func (rf *Raft) getFirstLog() Entry {
	return rf.log[0]
}

func (rf *Raft) isLogUpToDate(term, index int) bool {
	lastLog := rf.getLastLog()
	return term > lastLog.Term || (term == lastLog.Term && index >= lastLog.Index)
}

func (rf *Raft) matchLog(term, index int) bool {
	return index <= rf.getLastLog().Index &&
		rf.log[index-rf.getFirstLog().Index].Term == term
}

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func insertionSort(res []int) {
	l, r := 0, len(res)
	for i := l + 1; i < r; i++ {
		for j := i; j > l && res[j] < res[j-1]; j-- {
			res[j], res[j-1] = res[j-1], res[j]
		}
	}
}
