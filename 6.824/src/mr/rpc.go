package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

// task state
const (
	WorkerWait  = 0
	MapState    = 1
	ReduceState = 2
	FinishState = 3
)

type HeartBeatArgs struct{}

// HeartBeatReply 申请任务
type HeartBeatReply struct {
	WorkingType   int // task state
	MapContent    MapContent
	ReduceContent ReduceContent
}

// MapContent Map 类型上下文
type MapContent struct {
	FileName string // 要处理地文件名
	MapId    int    // MapId
	nReduce  int    // 将中间结果存成 nReduce 份
}

type MapFinishReply struct{}

// MapFinishArgs Map 结束后返参
type MapFinishArgs struct {
	MapId                 int
	intermediateFilenames []string
}

// ReduceContent Reduce 类型上下文
type ReduceContent struct {
	ReduceId              int
	intermediateFilenames []string // Map 处理后的所有中间状态文件名
}

type ReduceFinishReply struct{}

type ReduceFinishArgs struct {
	ReduceId       int
	ResultFilename string
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
