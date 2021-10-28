package mr

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strings"
	"time"
)

// KeyValue
// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose a reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// ByKey sorting by key of kvs
type ByKey []KeyValue

func (key ByKey) Len() int {
	return len(key)
}

func (key ByKey) Less(i, j int) bool {
	return key[i].Key < key[j].Key
}

func (key ByKey) Swap(i, j int) {
	key[i], key[j] = key[j], key[i]
}

// Worker
// main/mrworker.go calls this function.
// 通过 RPC 不断地向 Coordinator 申请 Task 并运行
// mapf 文件名 + 内容
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	// loop
	for {
		// 申请
		reply := callHeartBeat()
		fmt.Printf("Get a HeartBeat: %v", reply)
		// 策略判断
		switch reply.WorkingType {

		// state of Map
		case MapState:
			fmt.Printf("Doing a map state work: %v", reply.MapContent)
			mapContent := reply.MapContent
			// map start to open, read and close file
			fileName := mapContent.FileName
			file, err := os.Open(fileName)
			if err != nil {
				log.Fatalf("An unexpected error occurred opening map file %v", fileName)
			}

			context, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("An unexpected error occurred reading map file %v", fileName)
			}
			if err := file.Close(); err != nil {
				log.Fatalf("An unexpected error occurred closing map file %v", fileName)
			}

			// mapf
			kvs := mapf(fileName, string(context))
			sort.Sort(ByKey(kvs))

			// tmp 最终汇总
			var tmpFiles []*os.File
			var intermediateFilenames []string
			nReduce := mapContent.nReduce

			for i := 0; i < nReduce; i++ {
				tmpFile, err := ioutil.TempFile("data", fmt.Sprintf("mr-%d-%d-*.txt", mapContent.MapId, i))
				if err != nil {
					log.Fatal("An unexpected error occurred creating map tmp file")
				}
				tmpFiles = append(tmpFiles, tmpFile)
				intermediateFilenames = append(intermediateFilenames, tmpFile.Name())
			}

			for i := 0; i < len(kvs); i++ {
				kv := kvs[i]
				tmpFile := tmpFiles[ihash(kv.Key)%nReduce]
				fmt.Fprintf(tmpFile, "%v %v \n", kv.Key, kv.Value)
			}

			for i := 0; i < nReduce; i++ {
				tmpFile := tmpFiles[i]
				tmpFile.Close()
			}

			// map 向上汇报完成
			reportFinishMap(mapContent.MapId, intermediateFilenames)

		// state of Reduce
		case ReduceState:
			fmt.Printf("Doing a reduce state work: %v", reply.ReduceContent)
			reduceContent := reply.ReduceContent

			var kvs []KeyValue
			filenames := reduceContent.intermediateFilenames

			for i := 0; i < len(filenames); i++ {
				filename := filenames[i]
				// reduce start to open, read and close file
				file, err := os.Open("data/" + filename)
				if err != nil {
					log.Fatalf("An unexpected error occurred opening reduce file %v", filename)
				}

				content, err := ioutil.ReadAll(file)
				if err != nil {
					log.Fatalf("An unexpected error occurred reading map file %v", filename)
				}

				kvs = append(kvs, convertContent2Kvs(string(content))...)
			}

			sort.Sort(ByKey(kvs))
			outFile, err := ioutil.TempFile("data", fmt.Sprintf("mr-out-%d-*", reduceContent.ReduceId))
			if err != nil {
				log.Fatal("An unexpected error occurred creating reduce tmp file")
			}

			for i := 0; i < len(kvs); i++ {
				j := i + 1
				for j < len(kvs) && kvs[j].Key == kvs[i].Key {
					j++
				}
				var values []string
				for k := i; k < j; k++ {
					values = append(values, kvs[k].Value)
				}
				ret := reducef(kvs[i].Key, values)
				fmt.Fprintf(outFile, "%v %v \n", kvs[i].Key, ret)
				i = j
			}
			outFile.Close()
			// map 向上汇报完成
			reportFinishReduce(reduceContent.ReduceId, outFile.Name())

		case WorkerWait:
			fmt.Println("No task now, wait for 1s")
			time.Sleep(time.Second)
		case FinishState:
			fmt.Println("Receive finish signal")
			return
		default:
			fmt.Println("Something error occurred")
			return
		}
	}
}

// reportFinishReduce
func reportFinishReduce(reduceId int, filename string) {
	args := ReduceFinishArgs{
		reduceId, filename,
	}
	reply := ReduceFinishReply{}
	call("Coordinator.HandleReduceFinished", &args, &reply)
}

// convertContent2Kvs
func convertContent2Kvs(content string) []KeyValue {
	var kvs []KeyValue
	kv := strings.Split(content, "\n")
	for _, res := range kv {
		if len(res) == 0 {
			continue
		}
		tmp := strings.Split(res, " ")
		kvs = append(kvs, KeyValue{tmp[0], tmp[1]})
	}
	return kvs
}

// reportFinishMap
func reportFinishMap(mapId int, filenames []string) {
	args := MapFinishArgs{
		mapId, filenames,
	}
	reply := MapFinishReply{}
	call("Coordinator.HandleMapFinished", &args, &reply)
}

// callHeartBeat
// example function to show how to make an RPC call to the coordinator.
// the RPC argument and reply types are defined in rpc.go.
func callHeartBeat() *HeartBeatReply {
	args := HeartBeatArgs{}
	reply := HeartBeatReply{}
	call("Coordinator.ApplyForTask", &args, &reply)
	return &reply
}

// call
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcName string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
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
