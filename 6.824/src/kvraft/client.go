package kvraft

import "6.824/labrpc"
import "crypto/rand"
import "math/big"

type Clerk struct {
	servers   []*labrpc.ClientEnd
	leaderId  int64
	clientId  int64
	commandId int64 // (clientId, commandId) 唯一的确定一个请求
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigX, _ := rand.Int(rand.Reader, max)
	x := bigX.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	return &Clerk{
		servers:   servers,
		leaderId:  0,
		clientId:  nrand(),
		commandId: 0,
	}
}

// Get
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
func (ck *Clerk) Get(key string) string {
	return ck.Command(&CommandRequest{
		Key: key,
		Op:  OpGet,
	})
}

func (ck *Clerk) Put(key, value string) {
	ck.Command(&CommandRequest{
		Key:   key,
		Value: value,
		Op:    OpPut,
	})
}

func (ck *Clerk) Append(key, value string) {
	ck.Command(&CommandRequest{
		Key:   key,
		Value: value,
		Op:    OpAppend,
	})
}

// Command
// shared by Get Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
func (ck *Clerk) Command(request *CommandRequest) string {
	request.ClientId, request.CommandId = ck.clientId, ck.commandId
	for {
		var reply CommandReply
		// 发生错误处理
		if !ck.servers[ck.leaderId].Call("KVServer.Command", request, &reply) ||
			reply.Err == ErrWrongLeader || reply.Err == ErrTimeout {
			ck.leaderId = (ck.leaderId + 1) % int64(len(ck.servers))
			continue
		}
		ck.commandId++
		return reply.Value
	}
}
