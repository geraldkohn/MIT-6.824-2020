package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

type nodeType int16

const (
	node_mapper nodeType = iota
	node_reducer
)

//Map任务
type mapTask struct {
	taskId               int    //任务序号
	workerId             int    //节点序号
	filename             string //mapTask分配的文件
	intermediateFilename string //mapTask产生的中间文件
}

//reduce任务
type reduceTask struct {
	taskId        int    //任务序号
	workerId      int    //节点序号
	finalFilename string //reduceTask产生的最终文件
}

//map任务和reduce任务交接
type mapToReduceTask struct {
	mt mapTask
	rt reduceTask
}

//活跃的节点
type workerNode struct {
	nodeId   int
	nodeType nodeType
	ip       string
	port     string
}

type handleMapTaskMaster struct {
	//map正常运行节点
	mappers []workerNode
	//需要nMap个Map任务
	nMap int
	//map任务总共的文件
	totalFile []string
	//map任务中已经分配的文件
	assignedFile []string
	//map任务未分配的文件
	unassignedFile []string
	//正在运行的任务
	runningMapTasks []mapTask
	//已经完成的任务
	completedTasks []mapTask
	//是否都已经完成
	done bool
}

type handleAssignedMapTaskToReducerMaster struct {
	//已经完成的任务中已经分配给reduce节点的任务
	assignedMapTasks []mapToReduceTask
	//已经完成的任务中未分配给reduce节点的任务
	unsignedMapTasks []mapTask
	//是否reduce节点都已经完成任务
	done bool
}

type handleReduceTaskMaster struct {
	//reduce正常运行节点
	reducers []workerNode
	//需要nReduce个reduce任务
	nReduce int
	//正在运行的任务
	runningReduceTasks []reduceTask
	//已经完成的任务
	completedReduceTasks []reduceTask
	//是否当前所有正在运行的任务已经完成
	done bool
}

type Master struct {
	//活跃的节点
	workers           []workerNode
	mapMaster         handleMapTaskMaster
	mapToReduceMaster handleAssignedMapTaskToReducerMaster
	reduceMaster      handleReduceTaskMaster
	//是否任务全部结束
	done bool
}

// Your code here -- RPC handlers for the worker to call.

//mapper和reducer向master发送心跳
func (m *Master) Heartbeat(args *HeartbeatArgs, reply *HeartbeatReply) error {
	return nil
}

//新节点加入集群
func (m *Master) JoinCluster(args *JoinClusterArgs, reply *JoinClusterReply) error {
	return nil
}

//节点请求MapTask
func (m *Master) AskMapTask(args *AskMapTaskArgs, reply *AskMapTaskReply) error {
	return nil
}

//节点请求ReduceTask
func (m *Master) AskReduceTask(args *AskReduceTaskArgs, reply *AskReduceTaskReply) error {
	return nil
}

//节点做完MapTask, 告知数据的位置
func (m *Master) DoneMapTask(args *DoneMapTaskArgs, reply *DoneMapTaskReply) error {
	return nil
}

//节点做完ReduceTask, 告知数据的位置
func (m *Master) DoneReduceTask(args *DoneReduceTaskArgs, reply *DoneMapTaskReply) error {
	return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false

	// Your code here.

	return ret
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// Your code here.

	m.server()
	return &m
}
