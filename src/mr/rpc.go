package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

type HeartbeatArgs struct {
	NodeType nodeType
	Ip       string
	Port     string
	NodeId   int
}

type HeartbeatReply struct {
}

type JoinClusterArgs struct {
	Ip   string
	Port string
}

type JoinClusterReply struct {
	Agreement bool   //master节点是否同意加入集群
	nodeId    int    //节点序号
	Msg       string //返回的信息
}

type AskMapTaskArgs struct {
}

type AskMapTaskReply struct {
}

type AskReduceTaskArgs struct {
}

type AskReduceTaskReply struct {
}

type DoneMapTaskArgs struct {
}

type DoneMapTaskReply struct {
}

type DoneReduceTaskArgs struct {
}

type DoneReduceTaskReply struct {
}

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
