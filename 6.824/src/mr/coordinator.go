package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

func (c *Coordinator) schedule() {
	c.initMapPhase()
	for {
		select {
		case msg := <-c.heartbeatCh:
			if c.phase == Done {
				msg.response.JobType = Done
			} else if c.selectTask(msg.response) {
				switch c.phase {
				case Map:
					log.Printf("Coordinator: %v finished, start %v \n",
						Map, Reduce)
					c.initReducePhase()
					c.selectTask(msg.response)
				case Reduce:
					log.Printf("Coordinator: %v finished, Congratulations \n",
						Reduce)
					c.initCompletePhase()
					msg.response.JobType = Done
				case Done:
					panic(fmt.Sprintf("Coordinator: enter unexpected branch"))
				}
			}
			log.Printf("Coordinator: assigned a task %v to worker \n",
				msg.response)
			msg.ok <- struct{}{}
		case msg := <-c.reportCh:
			if msg.request.Phase == c.phase {
				log.Printf("Coordinator: Worker has executed task %v \n",
					msg.request)
				c.tasks[msg.request.Id].status = Finished
			}
			msg.ok <- struct{}{}
		}
	}
}

func (c *Coordinator) selectTask(response *HeartbeatResponse) bool {
	allFinished, hasNewJob := true, false
	for id, task := range c.tasks {
		switch task.status {
		case Init:
			allFinished, hasNewJob = false, true
			c.tasks[id].status, c.tasks[id].startTime = Working, time.Now()
			response.NReduce, response.Id = c.nReduce, id
			if c.phase == Map {
				response.JobType, response.MapFile = Map, c.files[id]
			} else {
				response.JobType, response.NMap = Reduce, c.nMap
			}
		case Working:
			allFinished = false
			if time.Now().Sub(task.startTime) > MaxTime {
				hasNewJob = true
				c.tasks[id].startTime = time.Now()
				response.NReduce, response.Id = c.nReduce, id
				if c.phase == Map {
					response.JobType, response.MapFile = Map, c.files[id]
				} else {
					response.JobType, response.NMap = Reduce, c.nMap
				}
			}
		case Finished:
		}
		if hasNewJob {
			break
		}
	}
	if !hasNewJob {
		response.JobType = Wait
	}
	return allFinished
}

func (c *Coordinator) initMapPhase() {
	c.phase = Map
	c.tasks = make([]Task, len(c.files))
	for index, file := range c.files {
		c.tasks[index] = Task{
			fileName: file,
			id:       index,
			status:   Init,
		}
	}
}

func (c *Coordinator) initReducePhase() {
	c.phase = Reduce
	c.tasks = make([]Task, c.nReduce)
	for i := 0; i < c.nReduce; i++ {
		c.tasks[i] = Task{
			id:     i,
			status: Init,
		}
	}
}

func (c *Coordinator) initCompletePhase() {
	c.phase = Done
	c.doneCh <- struct{}{}
}

func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	name := coordinatorSock()
	os.Remove(name)
	l, e := net.Listen("unix", name)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (c *Coordinator) Done() bool {
	<-c.doneCh
	return true
}

func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		files:       files,
		nReduce:     nReduce,
		nMap:        len(files),
		heartbeatCh: make(chan heartbeatMsg),
		reportCh:    make(chan reportMsg),
		doneCh:      make(chan struct{}, 1),
	}
	c.server()
	go c.schedule()
	return &c
}
