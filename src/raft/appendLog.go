package raft

func (rf *Raft) appendLog(command interface{}) (index int, term int) {
	if rf.sm.state != leader {
		return -1, -1
	}

	//leader append log
	rf.log = append(rf.log, logEntry{
		Term: rf.currentTerm,
		Command: command,
	})

	//leader distribute command to followers
	//下一次heartBeat就发送了

	//leader will commit if most of followers receive command and append log
	
	return
}
