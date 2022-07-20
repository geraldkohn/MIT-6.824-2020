package mr

import (
	"log"
	"testing"
)

func TestSet(t *testing.T) {

	task := newMapTaskInfo(0, "name-0", 10)

	set := newMapSet()
	set.Push(task)

	reply, err := set.Pop(task.GetTaskID())
	if err != nil {
		t.Error(err)
	}
	if reply.GetTaskID() != task.GetTaskID() {
		t.Error()
	}
	log.Println(reply.GetTaskType())

	if set.Has(task.GetTaskID()) {
		t.Error()
	}
}
