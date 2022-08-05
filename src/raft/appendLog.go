package raft

func (rf *Raft) appendLog(command interface{}) (index int, term int) {
	if rf.sm.state != leader {
		return -1, -1
	}

	//leader append log
	rf.rwLock(logEntriesMutex())
	index = len(rf.log)
	term = rf.currentTerm
	rf.log = append(rf.log, logEntry{
		Term:    rf.currentTerm,
		Command: command,
	})
	rf.rwUnlock(logEntriesMutex())

	//leader distribute command to followers
	//下一次heartBeat就发送了

	//leader will commit if most of followers receive command and append log
	for {
		select {
		case <-rf.sm.commitLog:
			if index <= rf.commitIndex {
				return
			} else {
				continue
			}
		}
	}
}

//节点serverId, 日志更新到了committedIndex(包括committedIndex)
func (rf *Raft) updateCommittedIndex(serverId int) {
	rf.rLock(logEntriesMutex())
	logLength := len(rf.log)
	rf.rUnlock(logEntriesMutex())

	newCommittedIndex := false
	for i := rf.commitIndex + 1; i <= logLength; i++ {
		//i下标的日志是否被过半数节点写入日志
		count := 0
		for _, m := range rf.matchIndex {
			if m >= i {
				count += 1
				if count > len(rf.peers)/2 {
					rf.commitIndex = i
					newCommittedIndex = true
					rf.DLog("update commit index: %d", i)
					break
				}
			}
		}
		//i下标的日志没有被过半数的节点写入日志
		if rf.commitIndex != i {
			break
		}
	}
	if newCommittedIndex {
		rf.sm.commitLog <- struct{}{}
	}
}
