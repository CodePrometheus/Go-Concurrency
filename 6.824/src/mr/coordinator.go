package mr

import (
	"fmt"
	"log"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

type Coordinator struct {
	mu      sync.Mutex
	nReduce int
	files   []string

	mapStates   []bool
	mapFinished []bool

	reduceStates   []bool
	reduceFinished []bool

	curState int
}

// ApplyForTask Task调度中心
func (c *Coordinator) ApplyForTask(reply *HeartBeatReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.curState {
	case MapState:
		content := MapContent{}
		for i := 0; i < len(c.mapStates); i++ {
			if c.mapStates[i] != true {
				content.MapId = i
				content.nReduce = c.nReduce
				content.FileName = c.files[i]
				c.mapStates[i] = true

				reply.WorkingType = MapState
				reply.MapContent = content
				fmt.Printf("Allocate a map work, MapContent: %v", reply.MapContent)
				// check map finished
				go c.asyncCheckMapFinished(i)
				return nil
			}
		}

	case ReduceState:
		content := ReduceContent{}
		for i := 0; i < len(c.reduceStates); i++ {
			if c.reduceStates[i] != true {
				content.ReduceId = i
				content.intermediateFilenames = c.getReduceFilenames(i)
				c.reduceStates[i] = true

				reply.WorkingType = ReduceState
				reply.ReduceContent = content
				fmt.Printf("Allocate a map work, ReduceContent: %v", reply.ReduceContent)
				// check reduce finished
				go c.asyncCheckReduceFinished(i)
				return nil
			}
		}

	case FinishState:
		fmt.Println("finish signal", reply)
		reply.WorkingType = FinishState
	default:
		fmt.Println("Something error occurred")
	}
	return nil
}

// asyncCheckReduceFinished
func (c *Coordinator) asyncCheckReduceFinished(reduceId int) {
	time.Sleep(time.Second * 10)
	c.mu.Lock()
	if c.reduceFinished[reduceId] != true {
		c.reduceFinished[reduceId] = false
	}
	c.mu.Unlock()
}

// getReduceFilenames
func (c *Coordinator) getReduceFilenames(reduceId int) []string {
	mapLen := len(c.mapStates)
	var filename []string
	for i := 0; i < mapLen; i++ {
		filename = append(filename, fmt.Sprintf("mr-%d-%d.txt", i, reduceId))
	}
	return filename
}

// asyncCheckMapFinished
func (c *Coordinator) asyncCheckMapFinished(mapId int) {
	time.Sleep(time.Second * 10)
	c.mu.Lock()
	if c.mapFinished[mapId] != true {
		c.mapFinished[mapId] = false
	}
	c.mu.Unlock()
}

func (c *Coordinator) HandleMapFinished(args *MapFinishArgs, reply *MapFinishReply) {
	// rename
	tmpFiles := args.intermediateFilenames
	for i := 0; i < len(tmpFiles); i++ {
		filename := tmpFiles[i]
		if err := os.Rename(filename, fmt.Sprintf("data/mr-%d-%d.txt", args.MapId, i)); err != nil {
			log.Fatalf("An unexpected error occurred rename map tmp file %v", filename)
		}
	}

	c.mu.Lock()
	c.mapFinished[args.MapId] = true
	c.mu.Unlock()

	if c.checkAllMapFinished() {
		fmt.Println("All Map State Finished")
		c.curState = ReduceState
	}
}

func (c *Coordinator) HandleReduceFinished(args *ReduceFinishArgs, reply *ReduceFinishReply) {
	tmpFilename := args.ResultFilename
	if err := os.Rename(tmpFilename, fmt.Sprintf("mr-out-%d", args.ReduceId)); err != nil {
		log.Fatalf("An unexpected error occurred rename reduce tmp file %v", tmpFilename)
	}
	c.mu.Lock()
	c.reduceFinished[args.ReduceId] = true
	c.mu.Unlock()
	if c.checkAllReduceFinished() {
		c.curState = FinishState
	}
}

func (c *Coordinator) checkAllReduceFinished() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < len(c.reduceFinished); i++ {
		if c.reduceFinished[i] == false {
			return false
		}
	}
	return true
}

func (c *Coordinator) checkAllMapFinished() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < len(c.mapFinished); i++ {
		if c.mapFinished[i] == false {
			return false
		}
	}
	return true
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockName := coordinatorSock()
	os.Remove(sockName)
	l, e := net.Listen("unix", sockName)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	go http.Serve(l, nil)
}

// Done
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.curState == FinishState
}

// MakeCoordinator
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	os.Mkdir("data", os.ModePerm)
	c := Coordinator{
		files:          files,
		nReduce:        nReduce,
		curState:       MapState,
		mapStates:      make([]bool, len(files)),
		mapFinished:    make([]bool, len(files)),
		reduceStates:   make([]bool, nReduce),
		reduceFinished: make([]bool, nReduce),
	}
	c.server()
	return &c
}
