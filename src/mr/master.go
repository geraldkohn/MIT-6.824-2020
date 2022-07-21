package mr

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type taskType int16

const (
	task_Map taskType = iota
	task_Reduce
	task_Wait
	task_End
)

type taskInfoInterface interface {
	GetTaskID() int
	GetTaskType() taskType
	SetBeginTime()
	UpdateHeartBeatTime()
	OutOfTime() bool
}

//map任务信息
type mapTaskInfo struct {
	beginTime     time.Time
	heartBeatTime time.Time

	mapIndex int    //m-mapIndex-[0,nReduce-1], n是这个文件应该由哪个reduce进程处理(中间文件)
	filename string //处理哪个文件
	nReduce  int
}

func newMapTaskInfo(mapIndex int, filename string, nReduce int) taskInfoInterface {
	return &mapTaskInfo{
		mapIndex: mapIndex,
		filename: filename,
		nReduce:  nReduce,
	}
}

func (info *mapTaskInfo) GetTaskID() int {
	return info.mapIndex
}

func (info *mapTaskInfo) GetTaskType() taskType {
	return task_Map
}

func (info *mapTaskInfo) SetBeginTime() {
	info.beginTime = time.Now()
}

func (info *mapTaskInfo) UpdateHeartBeatTime() {
	info.heartBeatTime = time.Now()
}

func (info *mapTaskInfo) OutOfTime() bool {
	return info.heartBeatTime.Sub(info.beginTime) > 10*time.Second
}

//reduce任务信息
type reduceTaskInfo struct {
	beginTime     time.Time
	heartBeatTime time.Time

	reduceIndex int //这是第几个reduce处理进程
	nMap        int //接收m-[0,nMap-1]-reduceIndex
}

func newReduceTaskInfo(reduceIndex int, nMap int) taskInfoInterface {
	return &reduceTaskInfo{
		reduceIndex: reduceIndex,
		nMap:        nMap,
	}
}

func (info *reduceTaskInfo) GetTaskID() int {
	return info.reduceIndex
}

func (info *reduceTaskInfo) GetTaskType() taskType {
	return task_Reduce
}

func (info *reduceTaskInfo) SetBeginTime() {
	info.beginTime = time.Now()
}

func (info *reduceTaskInfo) UpdateHeartBeatTime() {
	info.heartBeatTime = time.Now()
}

func (info *reduceTaskInfo) OutOfTime() bool {
	return info.heartBeatTime.Sub(info.beginTime) > 10*time.Second
}

//主节点
type Master struct {
	resource resourceMaster
}

//资源
type resourceMaster struct {
	mapTaskWaiting    *mapSet
	mapTaskRunning    *mapSet
	mapTaskDone       *mapSet
	reduceTaskWaiting *mapSet
	reduceTaskRunning *mapSet
	reduceTaskDone    *mapSet
}

func (m *resourceMaster) mapTaskWaitingToRunning() (taskInfoInterface, error) {
	task, err := m.mapTaskWaiting.PopRandomly()
	if err != nil {
		return nil, errors.New("mapTaskWaiting is empty")
	}
	task.SetBeginTime()
	task.UpdateHeartBeatTime()
	m.mapTaskRunning.Push(task)
	return task, nil
}

func (m *resourceMaster) mapTaskRunningToDone(mapIndex int) error {
	task, err := m.mapTaskRunning.Pop(mapIndex)
	if err != nil {
		return errors.New("mapTaskRunning do not have special mapTask")
	}
	m.mapTaskDone.Push(task)
	return nil
}

func (m *resourceMaster) mapTaskDoneToWaiting(mapIndex int) error {
	task, err := m.mapTaskDone.Pop(mapIndex)
	if err != nil {
		return errors.New("mapTaskDone do not have special mapTask")
	}
	m.mapTaskWaiting.Push(task)
	return nil
}

func (m *resourceMaster) mapTaskRunningToWaiting(mapIndex int) error {
	task, err := m.mapTaskRunning.Pop(mapIndex)
	if err != nil {
		return errors.New("mapTaskRunning do not have special mapTask")
	}
	m.mapTaskWaiting.Push(task)
	return nil
}

func (m *resourceMaster) reduceTaskWaitingToRunning() (taskInfoInterface, error) {
	task, err := m.reduceTaskWaiting.PopRandomly()
	if err != nil {
		return nil, errors.New("reduceTaskWaiting is empty")
	}
	task.SetBeginTime()
	task.UpdateHeartBeatTime()
	m.reduceTaskRunning.Push(task)
	return task, nil
}

func (m *resourceMaster) reduceTaskRunningToDone(reduceIndex int) error {
	task, err := m.reduceTaskRunning.Pop(reduceIndex)
	if err != nil {
		return errors.New("reduceTaskRunning do not have special mapTask")
	}
	m.reduceTaskDone.Push(task)
	return nil
}

func (m *resourceMaster) reduceTaskDoneToWaiting(reduceIndex int) error {
	task, err := m.reduceTaskDone.Pop(reduceIndex)
	if err != nil {
		return errors.New("reduceTaskDone do not have special mapTask")
	}
	m.reduceTaskWaiting.Push(task)
	return nil
}

func (m *resourceMaster) reduceTaskRunningToWaiting(reduceIndex int) error {
	task, err := m.reduceTaskRunning.Pop(reduceIndex)
	if err != nil {
		return errors.New("reduceTaskRunning do not have special mapTask")
	}
	m.reduceTaskWaiting.Push(task)
	return nil
}

func (m *resourceMaster) startMapTask(filename []string, nReduce int) {
	cnt := 0
	for _, file := range filename {
		task := newMapTaskInfo(cnt, file, nReduce)
		m.mapTaskWaiting.Push(task)
		cnt++
	}
}

func (m *resourceMaster) startReduceTask(filename []string, nReduce int) {
	nMap := len(filename)
	for i := 0; i < nReduce; i++ {
		task := newReduceTaskInfo(i, nMap)
		m.reduceTaskWaiting.Push(task)
	}
}

//巡查正在运行的任务, 看看有没有超时的, 超时的节点放到等待集合
func (m *resourceMaster) scanNode() {
	for {
		if m.mapTaskRunning.size() != 0 {
			for k, v := range m.mapTaskRunning.Bool {
				if v.OutOfTime() {
					err := m.mapTaskRunningToWaiting(k)
					if err != nil {
						log.Println("scan node error, " + err.Error())
					}
				}
			}
		}

		if m.reduceTaskRunning.size() != 0 {
			for k, v := range m.reduceTaskRunning.Bool {
				if v.OutOfTime() {
					err := m.reduceTaskRunningToWaiting(k)
					if err != nil {
						log.Println("scan node error, " + err.Error())
					}
				}
			}
		}

		//没间隔两秒扫描一遍节点
		time.Sleep(2 * time.Second)
	}

}

/** rpc调用 */

//节点请求任务
func (m *Master) AskForTask(args *AskForTaskArgs, reply *AskForTaskReply) error {
	//如果map任务已经结束, 开始分配task任务
	if m.resource.mapTaskWaiting.size() == 0 && m.resource.mapTaskRunning.size() == 0 {
		task, err := m.resource.reduceTaskWaitingToRunning()
		if err != nil {
			return err
		}
		reply.taskInfo = task
	} else {
		task, err := m.resource.mapTaskWaitingToRunning()
		if err != nil {
			return err
		}
		reply.taskInfo = task
	}
	return nil
}

//节点告知任务已经完成
func (m *Master) TaskDone(args *TaskDoneArgs, reply *TaskDoneReply) error {
	task := args.taskInfo
	switch task.GetTaskType() {
	case task_Reduce:
		err := m.resource.reduceTaskRunningToDone(task.GetTaskID())
		if err != nil {
			return err
		}
	case task_Map:
		err := m.resource.mapTaskRunningToDone(task.GetTaskID())
		if err != nil {
			return err
		}
	default:
		return errors.New("task type error")
	}
	return nil
}

//节点发送心跳
func (m *Master) HeartBeat(args *HeartBeatArgs, reply *HeartBeatReply) error {
	task := args.taskInfo
	switch task.GetTaskType() {
	case task_Map:
		task, err := m.resource.mapTaskRunning.Pop(task.GetTaskID())
		if err != nil {
			return err
		}
		task.UpdateHeartBeatTime()
		m.resource.mapTaskRunning.Push(task)
	case task_Reduce:
		task, err := m.resource.reduceTaskRunning.Pop(task.GetTaskID())
		if err != nil {
			return err
		}
		task.UpdateHeartBeatTime()
		m.resource.reduceTaskRunning.Push(task)
	default:
		return errors.New("task type error")
	}
	return nil
}

// Your code here -- RPC handlers for the worker to call.

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
	if m.resource.mapTaskRunning.size() == 0 && m.resource.mapTaskWaiting.size() == 0 && m.resource.reduceTaskRunning.size() == 0 && m.resource.reduceTaskWaiting.size() == 0 {
		ret = true
	}

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
	m.resource.startMapTask(files, nReduce)
	m.resource.startReduceTask(files, nReduce)
	m.resource.scanNode()

	m.server()
	return &m
}
