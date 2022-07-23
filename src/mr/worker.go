package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

type kvArray []KeyValue

func (kva kvArray) Len() int           { return len(kva) }
func (kva kvArray) Swap(i, j int)      { kva[i], kva[j] = kva[j], kva[i] }
func (kva kvArray) Less(i, j int) bool { return kva[i].Key < kva[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

type Node struct {
	mapf    func(string, string) []KeyValue
	reducef func(string, []string) string

	//任务
	taskInfo taskInfoInterface
}

/** 处理Map任务 */

// 生成中间k, v值
func (node *Node) makeIntermediateForMapTask(filePath string) []KeyValue {
	//打开文件
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("can not open %v", filePath)
	}
	//读取文件
	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("can not read file %v", filePath)
	}
	file.Close()
	//生成k, v数组
	kva := node.mapf(filePath, string(content))
	return kva
}

// 生成文件mr-mapTaskID-[0,nReduce-1]
func (node *Node) writeToFilesForMapTask(mapTaskID int, nReduce int, intermediate []KeyValue) {
	kvas := make([][]KeyValue, nReduce)
	for i := 0; i < nReduce; i++ {
		kvas[i] = make([]KeyValue, 0)
	}
	for _, kv := range intermediate {
		index := ihash(kv.Key) % nReduce
		kvas[index] = append(kvas[index], kv)
	}
	for i := 0; i < nReduce; i++ {
		// use this as equivalent
		tempFile, err := os.CreateTemp(".", "mrtemp")
		if err != nil {
			log.Fatal(err)
		}
		encoder := json.NewEncoder(tempFile)
		for _, kv := range kvas[i] {
			err := encoder.Encode(&kv)
			if err != nil {
				log.Fatal(err)
			}
		}
		outname := fmt.Sprintf("mr-%v-%v", mapTaskID, i)
		err = os.Rename(tempFile.Name(), outname)
		if err != nil {
			log.Println(err)
		}
	}
}

// 执行Map任务
func (node *Node) executeMapTask(reply *AskForTaskReply) {
	task, ok := reply.taskInfo.(*mapTaskInfo)
	if !ok {
		log.Fatal("类型断言出错")
	}

	log.Println("开始执行Map任务")
	intermediate := node.makeIntermediateForMapTask(task.filename)
	node.writeToFilesForMapTask(task.mapIndex, task.nReduce, intermediate)
	node.rpcTaskDone()
}

/** 处理Reduce任务 */

// 从map生成的中间文件中读取k, v对
func (node *Node) readFilesForReduceTask(mapTaskIndex int, reduceTaskIndex int) []KeyValue {
	filename := fmt.Sprintf("mr-%v-%v", mapTaskIndex, reduceTaskIndex)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("can not open file: %v\n", filename)
	}
	decoder := json.NewDecoder(file)
	kva := make([]KeyValue, 0)
	for {
		var kv KeyValue
		err = decoder.Decode(&kv)
		if err != nil {
			break
		}
		kva = append(kva, kv)
	}
	file.Close()
	return kva
}

func (node *Node) writeToFileForReduceTask(intermediate []KeyValue, taskID int) {
	outname := fmt.Sprintf("mr-out-%v", taskID)
	tempFile, err := os.CreateTemp(".", "mrtemp")
	if err != nil {
		log.Fatal("can not create temp file: ", err)
	}
	sort.Sort(kvArray(intermediate))
	for i := 0; i < len(intermediate); {
		//找出key相同的k, v记录
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		//相同的key的value写入到values中, 传递给reduce函数
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := node.reducef(intermediate[i].Key, values)

		//将正确的形式写入到tempFile里(每一个key一行)
		fmt.Println(tempFile, "%v %v\n", intermediate[i].Key, output)
		i = j
	}
	tempFile.Close()
	err = os.Rename(tempFile.Name(), outname)
	if err != nil {
		log.Printf("重命名文件失败 %v\n", outname)
	}
}

func (node *Node) executeReduceTask(reply *AskForTaskReply) {
	task, ok := reply.taskInfo.(*reduceTaskInfo)
	if !ok {
		log.Fatalf("类型断言出错")
	}

	intermediate := make([]KeyValue, 0)
	for i := 0; i < task.nMap; i++ {
		intermediate = append(intermediate, node.readFilesForReduceTask(i, task.reduceIndex)...)
	}

	node.writeToFileForReduceTask(intermediate, task.reduceIndex)

	node.rpcTaskDone()
}

func (node *Node) rpcTaskDone() {
	args := TaskDoneArgs{}
	args.taskInfo = node.taskInfo
	reply := TaskDoneReply{}
	call("Master.TaskDone", &args, &reply)
}

func (node *Node) rpcAskForTask() *AskForTaskReply {
	args := AskForTaskArgs{}
	reply := AskForTaskReply{}
	call("Master.AskForTask", &args, &reply)
	return &reply
}

func (node *Node) rpcHeartBeat() {
	args := HeartBeatArgs{
		taskInfo: node.taskInfo,
	}
	reply := HeartBeatReply{}
	for {
		call("Master.HeartBeat", &args, &reply)
		time.Sleep(2 * time.Second)
	}
}

func (node *Node) process() {
	for {
		reply := node.rpcAskForTask()
		switch reply.taskType {
		case task_Map:
			node.executeMapTask(reply)
			node.rpcHeartBeat()
		case task_Reduce:
			node.executeReduceTask(reply)
			node.rpcHeartBeat()
		case task_Wait:
			time.Sleep(2 * time.Second)
		case task_End:
			return
		}
	}
}

func newNode(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) *Node {
	return &Node{
		mapf:    mapf,
		reducef: reducef,
	}
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	// uncomment to send the Example RPC to the master.
	// CallExample()

	node := newNode(mapf, reducef)
	node.process()
}

//
// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Master.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
