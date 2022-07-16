package mr

type uidType int32

const (
	uid_task uidType = iota
	uid_node
)

var uidCount = struct {
	taskId int
	nodeId int
}{0, 0}

//task, mapper, reducer
func newUid(u uidType) int {
	switch u {
	case uid_task:
		uidCount.taskId++
		return uidCount.taskId
	case uid_node:
		uidCount.nodeId++
		return uidCount.nodeId
	default:
		panic("uid pass bad parameters")
	}
}
