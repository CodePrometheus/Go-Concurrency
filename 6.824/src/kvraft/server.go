package kvraft

import (
	"6.824/labgob"
	"6.824/labrpc"
	"6.824/raft"
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type KVServer struct {
	mu      sync.RWMutex
	rf      *raft.Raft
	dead    int32
	applyCh chan raft.ApplyMsg

	maxRaftState int // snapshot if log grows this big
	lastApplied  int

	db             db
	lastOperations map[int64]OperationContext // 幂等性保证，记录上一次commandId及clientId对应的reply
	waitApplyChan  map[int]chan *CommandReply // 唤醒客户端线程
}

func (kv *KVServer) Command(request *CommandRequest, reply *CommandReply) {
	defer DPrintf("{ Node %v } processes CommandRequest %v with CommandReply %v",
		kv.rf.Me(), request, reply)

	kv.mu.RLock()
	if request.Op != OpGet && kv.isDuplicateRequest(request.ClientId, request.CommandId) {
		lastReply := kv.lastOperations[request.ClientId].LastReply
		reply.Value, reply.Err = lastReply.Value, lastReply.Err
		kv.mu.RUnlock()
		return
	}
	kv.mu.RUnlock()

	idx, _, isLeader := kv.rf.Start(Command{request})
	if !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	kv.mu.Lock()
	waitChan := kv.getWaitApplyChan(idx)
	kv.mu.Unlock()

	select {
	case res := <-waitChan:
		reply.Value, reply.Err = res.Value, res.Err
	case <-time.After(ExecuteTimeout):
		reply.Err = ErrTimeout
	}

	go func() {
		kv.mu.Lock()
		kv.removeOutdatedNotifyChan(idx)
		kv.mu.Unlock()
	}()
}

func (kv *KVServer) removeOutdatedNotifyChan(idx int) {
	delete(kv.waitApplyChan, idx)
}

func (kv *KVServer) getWaitApplyChan(idx int) chan *CommandReply {
	if _, ok := kv.waitApplyChan[idx]; !ok {
		kv.waitApplyChan[idx] = make(chan *CommandReply, 1)
	}
	return kv.waitApplyChan[idx]
}

func (kv *KVServer) applier() {
	for !kv.killed() {
		select {
		case msg := <-kv.applyCh:
			DPrintf("{ Node %v } tries to apply msg %v", kv.rf.Me(), msg)
			if msg.CommandValid {
				kv.mu.Lock()
				if msg.CommandIndex <= kv.lastApplied {
					DPrintf("{ Node %v } discards outdated msg %v because a newer snapshot which "+
						"lastApplied is %v has been restored", kv.rf.Me(), msg, kv.lastApplied)
					kv.mu.Unlock()
					continue
				}
				kv.lastApplied = msg.CommandIndex

				var reply *CommandReply
				command := msg.Command.(Command)
				if command.Op != OpGet && kv.isDuplicateRequest(command.ClientId, command.CommandId) {
					DPrintf("{ Node %v } doesn't apply duplicated message %v to"+
						" stateMachine because maxAppliedCommandId is %v for client %v",
						kv.rf.Me(), msg, kv.lastOperations[command.ClientId], command.ClientId)
					reply = kv.lastOperations[command.ClientId].LastReply
				} else {
					// apply log
					reply = kv.applyLogToStateMachine(command)
					if command.Op != OpGet {
						kv.lastOperations[command.ClientId] = OperationContext{
							command.CommandId, reply,
						}
					}
				}

				if currentTerm, isLeader := kv.rf.GetState(); isLeader && msg.CommandTerm == currentTerm {
					waitApplyChan := kv.getWaitApplyChan(msg.CommandIndex)
					waitApplyChan <- reply
				}

				needSnapshot := kv.needSnapshot()
				if needSnapshot {
					kv.takeSnapshot(msg.CommandIndex)
				}

				kv.mu.Unlock()
			} else if msg.SnapshotValid {
				kv.mu.Lock()
				if kv.rf.CondInstallSnapshot(msg.SnapshotTerm, msg.SnapshotIndex, msg.Snapshot) {
					kv.restoreSnapshot(msg.Snapshot)
					kv.lastApplied = msg.SnapshotIndex
				}
				kv.mu.Unlock()
			} else {
				panic(fmt.Sprintf("unexpected Message %v", msg))
			}
		}
	}
}

func (kv *KVServer) needSnapshot() bool {
	return kv.maxRaftState != -1 && kv.rf.GetRaftStateSize() >= kv.maxRaftState
}

func (kv *KVServer) takeSnapshot(idx int) {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(kv.db)
	e.Encode(kv.lastOperations)
	kv.rf.Snapshot(idx, w.Bytes())
}

func (kv *KVServer) restoreSnapshot(snapshot []byte) {
	if snapshot == nil || len(snapshot) == 0 {
		return
	}
	r := bytes.NewBuffer(snapshot)
	d := labgob.NewDecoder(r)
	var db DbKV
	var lastOperations map[int64]OperationContext
	if d.Decode(&db) != nil ||
		d.Decode(&lastOperations) != nil {
		DPrintf("{ Node %v } restores snapshot failed", kv.rf.Me())
	}
	kv.db, kv.lastOperations = &db, lastOperations
}

func (kv *KVServer) applyLogToStateMachine(command Command) *CommandReply {
	var value string
	var err Err
	switch command.Op {
	case OpPut:
		err = kv.db.Put(command.Key, command.Value)
	case OpGet:
		value, err = kv.db.Get(command.Key)
	case OpAppend:
		err = kv.db.Append(command.Key, command.Value)
	}
	return &CommandReply{err, value}
}

func (kv *KVServer) isDuplicateRequest(clientId int64, commandId int64) bool {
	operationContext, ok := kv.lastOperations[clientId]
	return ok && commandId <= operationContext.MaxAppliedCommandId
}

// Kill
// the tester calls Kill() when a KVServer instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.
func (kv *KVServer) Kill() {
	DPrintf("{ Node %v } has been killed", kv.rf.Me())
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
}

func (kv *KVServer) killed() bool {
	return atomic.LoadInt32(&kv.dead) == 1
}

// StartKVServer
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots through the underlying Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxRaftState int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(Command{})
	applyCh := make(chan raft.ApplyMsg)

	kv := &KVServer{
		dead:           0,
		lastApplied:    0,
		maxRaftState:   maxRaftState,
		db:             NewDbKV(),
		lastOperations: make(map[int64]OperationContext),
		waitApplyChan:  make(map[int]chan *CommandReply),
		applyCh:        applyCh,
		rf:             raft.Make(servers, me, persister, applyCh),
	}

	// commit 到状态机
	go kv.applier()

	DPrintf("{ Node %v } has started", kv.rf.Me())
	return kv
}
