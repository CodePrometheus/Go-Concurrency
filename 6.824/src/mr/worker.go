package mr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type KeyValue struct {
	Key   string
	Value string
}

func iHash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func Worker(mapF func(string, string) []KeyValue,
	reduceF func(string, []string) string) {
	for {
		response := doHeartbeat()
		log.Printf("Worker: receive coordinator's heartbeat %v \n", response)
		switch response.JobType {
		case Map:
			doMapTask(mapF, response)
		case Reduce:
			doReduceTask(reduceF, response)
		case Wait:
			time.Sleep(1 * time.Second)
		case Done:
			return
		default:
			panic(fmt.Sprintf("unexpected jobType %v", response.JobType))
		}
	}
}

func doMapTask(mapF func(string, string) []KeyValue, response *HeartbeatResponse) {
	fileName := response.MapFile
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("cannot open %v", fileName)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", fileName)
	}
	file.Close()
	kva := mapF(fileName, string(content))
	intermediates := make([][]KeyValue, response.NReduce)
	for _, kv := range kva {
		index := iHash(kv.Key) % response.NReduce
		intermediates[index] = append(intermediates[index], kv)
	}
	var wg sync.WaitGroup
	for reduceId, intermediate := range intermediates {
		wg.Add(1)
		go func(reduceId int, intermediate []KeyValue) {
			defer wg.Done()
			intermediateFile := mapOutFileName(response.Id, reduceId)
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			for _, kv := range intermediate {
				err := enc.Encode(&kv)
				if err != nil {
					log.Fatalf("cannot encode json %v", kv.Key)
				}
			}
			renameFilename(intermediateFile, &buf)
		}(reduceId, intermediate)
	}
	wg.Wait()
	doReport(response.Id, Map)
}

func doReduceTask(reduceF func(string, []string) string, response *HeartbeatResponse) {
	var kva []KeyValue
	for i := 0; i < response.NMap; i++ {
		fileName := mapOutFileName(i, response.Id)
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("cannot open %v", fileName)
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kva = append(kva, kv)
		}
		file.Close()
	}
	results := make(map[string][]string)
	for _, kv := range kva {
		results[kv.Key] = append(results[kv.Key], kv.Value)
	}
	var buf bytes.Buffer
	for key, values := range results {
		output := reduceF(key, values)
		fmt.Fprintf(&buf, "%v %v\n", key, output)
	}
	renameFilename(reduceOutFileName(response.Id), &buf)
	doReport(response.Id, Reduce)
}

func doHeartbeat() *HeartbeatResponse {
	response := HeartbeatResponse{}
	call("Coordinator.Heartbeat",
		&HeartbeatRequest{},
		&response)
	return &response
}

func doReport(id int, phase string) {
	call("Coordinator.Report",
		&ReportRequest{id, phase},
		&ReportResponse{})
}

func call(rpcName string, args interface{}, reply interface{}) bool {
	sockName := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockName)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()
	err = c.Call(rpcName, args, reply)
	if err == nil {
		return true
	}
	fmt.Println(err)
	return false
}
