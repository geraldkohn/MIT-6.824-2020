package mr

// import (
// 	"errors"
// 	"sync"
// )

// type listNode struct {
// 	data interface{}
// 	next *listNode
// 	prev *listNode
// }

// func (node *listNode) addBefore(data interface{}) {
// 	prev := node.prev

// 	newNode := &listNode{
// 		data: data,
// 		next: node,
// 		prev: prev,
// 	}

// 	prev.next = newNode
// 	node.prev = newNode
// }

// func (node *listNode) addAfter(data interface{}) {
// 	next := node.next

// 	newNode := &listNode{
// 		data: data,
// 		next: next,
// 		prev: node,
// 	}

// 	next.prev = newNode
// 	node.next = newNode
// }

// func (node *listNode) removeBefore() {
// 	prev := node.prev.prev
// 	node.prev = prev
// 	prev.next = node
// }

// func (node *listNode) removeAfter() {
// 	next := node.next.next
// 	node.next = next
// 	next.prev = node
// }

// type linkedList struct {
// 	head  *listNode
// 	count int
// }

// func newLinkedList() *linkedList {
// 	list := &linkedList{}
// 	list.count = 0
// 	list.head.next = list.head
// 	list.head.prev = list.head
// 	return list
// }

// func (list *linkedList) pushFront(data interface{}) {
// 	list.head.addAfter(data)
// 	list.count++
// }

// func (list *linkedList) pushBack(data interface{}) {
// 	list.head.addBefore(data)
// 	list.count++
// }

// func (list *linkedList) peekFront() (interface{}, error) {
// 	if list.count == 0 {
// 		return nil, errors.New("peeking empty list")
// 	}
// 	return list.head.next.data, nil
// }

// func (list *linkedList) peekBack() (interface{}, error) {
// 	if list.count == 0 {
// 		return nil, errors.New("peeking empty list")
// 	}
// 	return list.head.prev.data, nil
// }

// func (list *linkedList) popFront() (interface{}, error) {
// 	if list.count == 0 {
// 		return nil, errors.New("popping empty list")
// 	}
// 	data := list.head.next.data
// 	list.head.removeAfter()
// 	list.count--
// 	return data, nil
// }

// func (list *linkedList) popBack() (interface{}, error) {
// 	if list.count == 0 {
// 		return nil, errors.New("popping empty list")
// 	}
// 	data := list.head.prev.data
// 	list.head.removeBefore()
// 	list.count--
// 	return data, nil
// }

// func (list *linkedList) size() int {
// 	return list.count
// }

// //blockQueue 队列
// type blockQueue struct {
// 	list *linkedList
// 	cond *sync.Cond
// }

// func newBlockQueue() *blockQueue {
// 	queue := &blockQueue{}
// 	queue.list = newLinkedList()
// 	queue.cond = sync.NewCond(&sync.Mutex{})
// 	return queue
// }

// func (queue *blockQueue) pushFront(data interface{}) {
// 	queue.cond.L.Lock()
// 	queue.list.pushFront(data)
// 	queue.cond.L.Unlock()
// 	queue.cond.Broadcast()
// }

// func (queue *blockQueue) pushBack(data interface{}) {
// 	queue.cond.L.Lock()
// 	queue.list.pushBack(data)
// 	queue.cond.L.Unlock()
// 	queue.cond.Broadcast()
// }

// func (queue *blockQueue) peekFront() (interface{}, error) {
// 	queue.cond.L.Lock()
// 	for queue.list.count == 0 {
// 		queue.cond.Wait()
// 	}
// 	data, err := queue.list.peekFront()
// 	queue.cond.L.Unlock()
// 	return data, err
// }

// func (queue *blockQueue) peekBack() (interface{}, error) {
// 	queue.cond.L.Lock()
// 	for queue.list.count == 0 {
// 		queue.cond.Wait()
// 	}
// 	data, err := queue.list.peekBack()
// 	queue.cond.L.Unlock()
// 	return data, err
// }

// func (queue *blockQueue) popFront() (interface{}, error) {
// 	queue.cond.L.Lock()
// 	data, err := queue.list.popFront()
// 	queue.cond.L.Unlock()
// 	return data, err
// }

// func (queue *blockQueue) popBack() (interface{}, error) {
// 	queue.cond.L.Lock()
// 	data, err := queue.list.popBack()
// 	queue.cond.L.Unlock()
// 	return data, err
// }
