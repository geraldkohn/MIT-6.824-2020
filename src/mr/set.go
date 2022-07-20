package mr

import (
	"errors"
)

type mapSet struct {
	Bool  map[int]taskInfoInterface
	count int
}

func newMapSet() *mapSet {
	m := mapSet{}
	m.Bool = make(map[int]taskInfoInterface)
	m.count = 0
	return &m
}

func (m *mapSet) Push(task taskInfoInterface) {
	m.Bool[task.GetTaskID()] = task
	m.count++
}

func (m *mapSet) Has(taskID int) bool {
	_, ok := m.Bool[taskID]
	return ok
}

func (m *mapSet) Get(taskID int) (taskInfoInterface, error) {
	task, ok := m.Bool[taskID]
	if ok {
		return task, nil
	} else {
		return nil, errors.New("no task exists")
	}
}

func (m *mapSet) PopRandomly() (taskInfoInterface, error) {
	if m.count == 0 {
		return nil, errors.New("no task exists")
	}
	for _, task := range m.Bool {
		return task, nil
	}
	return nil, errors.New("no task exists")
}

func (m *mapSet) Pop(taskID int) (taskInfoInterface, error) {
	_, ok := m.Bool[taskID]
	if ok {
		task := m.Bool[taskID]
		delete(m.Bool, taskID)
		m.count--
		return task, nil
	} else {
		return nil, errors.New("no task exists")
	}
}

func (m *mapSet) size() int {
	return m.count
}
