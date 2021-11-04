package mr

import (
	"fmt"
	"os"
	"time"
)
import "strconv"

const (
	MaxTime = time.Second * 10
)

type Task struct {
	fileName  string
	id        int
	startTime time.Time
	status    TaskStatus
}

type Coordinator struct {
	files       []string
	nReduce     int
	nMap        int
	phase       string
	tasks       []Task
	heartbeatCh chan heartbeatMsg
	reportCh    chan reportMsg
	doneCh      chan struct{}
}

type heartbeatMsg struct {
	response *HeartbeatResponse
	ok       chan struct{}
}

type reportMsg struct {
	request *ReportRequest
	ok      chan struct{}
}

func (c *Coordinator) Heartbeat(request *HeartbeatRequest, response *HeartbeatResponse) error {
	msg := heartbeatMsg{response, make(chan struct{})}
	c.heartbeatCh <- msg
	<-msg.ok
	return nil
}

func (c *Coordinator) Report(request *ReportRequest, response *ReportResponse) error {
	msg := reportMsg{request, make(chan struct{})}
	c.reportCh <- msg
	<-msg.ok
	return nil
}

type HeartbeatRequest struct {
}

type HeartbeatResponse struct {
	MapFile string
	JobType string
	NReduce int
	NMap    int
	Id      int
}

func (response HeartbeatResponse) String() string {
	switch response.JobType {
	case Map:
		return fmt.Sprintf("{JobType: %v, FilePath: %v, Id: %v, NReduce: %v}",
			response.JobType, response.MapFile, response.Id, response.NReduce)
	case Reduce:
		return fmt.Sprintf("{JobType: %v, Id: %v, NMap: %v, NReduce: %v}", response.JobType, response.Id, response.NMap, response.NReduce)
	case Done, Wait:
		return fmt.Sprintf("{JobType: %v}", response.JobType)
	}
	panic(fmt.Sprintf("unexpected JobType %v", response.JobType))
}

const (
	Map    = "Map"
	Reduce = "Reduce"
	Done   = "Done"
	Wait   = "Wait"
)

type TaskStatus uint8

const (
	Init TaskStatus = iota
	Working
	Finished
)

type ReportRequest struct {
	Id    int
	Phase string
}

func (request ReportRequest) String() string {
	return fmt.Sprintf("{Id:%v,SchedulePhase:%v}", request.Id, request.Phase)
}

type ReportResponse struct {
}

func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
