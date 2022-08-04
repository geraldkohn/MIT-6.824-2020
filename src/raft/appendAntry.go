package raft

type AppendEntriesArgs struct {
	term         int //leader's term
	leaderId     int //so follower can redirect clients
	prevLogIndex int //index of log entry immediately preceding new ones
	prevLogTerm  int //term of preLogIndex entry
	entries      []logEntry
	leaderCommit int
}

type AppendEntriesReply struct {
	term    int
	success bool
	xTerm   int //这个是Follower中与Leader冲突的Log对应的任期号。在之前（7.1）有介绍Leader会在prevLogTerm中带上本地Log记录中，前一条Log的任期号。如果Follower在对应位置的任期号不匹配，它会拒绝Leader的AppendEntries消息，并将自己的任期号放在XTerm中。如果Follower在对应位置没有Log，那么这里会返回 -1
	xIndex  int //这个是Follower中，对应任期号为XTerm的第一条Log条目的槽位号
	xLen    int //如果Follower在对应位置没有Log，那么XTerm会返回-1，XLen表示空白的Log槽位数
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	if rf.sm.killed {
		return
	}

	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.DebugLog("get entries: %+v", args)

	//判断任期
	if rf.sm.state == candidate {
		if args.term >= rf.currentTerm {
			rf.stateConverter(candidate, follower)
		}
		reply.term = rf.currentTerm
		reply.success = false
		return
	}

	//判断任期
	if rf.sm.state == leader {
		if args.term >= rf.currentTerm {
			//需要状态转换
			rf.stateConverter(leader, follower)
		}
		reply.term = rf.currentTerm
		reply.success = false
		return
	}

	//当前节点term比leaderterm高
	if rf.currentTerm > args.term {
		reply.term = rf.currentTerm
		reply.success = false
		return
	}

	/** 快速恢复日志 */

	//当前位置没有日志
	//日志从下标为1开始, 但是日志最后的下标仍然为 len() - 1
	if len(rf.log)-1 < args.prevLogIndex {
		reply.xTerm = -1
		reply.xLen = args.prevLogIndex - len(rf.log)
		reply.success = false
		reply.term = rf.currentTerm
	}

	if rf.log[args.prevLogIndex].Term == args.prevLogTerm {
		//index同term同, 之前都相同
		reply.term = rf.currentTerm
		reply.success = true
		//更新节点日志
		rf.log = rf.log[:args.prevLogIndex+1]
		rf.log = append(rf.log, args.entries...)
		rf.persist()
		return
	} else {
		//index同Term不同, 需要继续向前找
		reply.xTerm = rf.log[args.prevLogIndex].Term
		xIndex := 0
		var i int
		for i = args.prevLogIndex; i > 1; i-- {
			if rf.log[i].Term != rf.log[args.prevLogIndex].Term {
				xIndex = i + 1
				break
			}
		}
		if i == 0 {
			xIndex = 1
		}
		reply.xIndex = xIndex
		reply.success = false
		reply.term = rf.currentTerm
		return
	}

}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntriesToEachPeers() {
	if rf.sm.killed {
		return
	}

	//获得返回值
	// replyCh := make(chan AppendEntriesReply, len(rf.peers))
	//通知主进程退出, 因为有term更高的节点
	exit := make(chan struct{}, 1)

	//开始需要通知各个节点
	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		go func(exit chan struct{}, i int) {
			args := &AppendEntriesArgs{
				term:         rf.currentTerm,
				leaderId:     rf.me,
				prevLogIndex: rf.nextIndex[i] - 1,
				prevLogTerm:  rf.log[rf.nextIndex[i]-1].Term,
				entries:      nil,
				leaderCommit: rf.commitIndex,
			}
			//上一次heartBeat到这一次heartBeat有命令
			if rf.nextIndex[i] < len(rf.log) {
				//防止数组越界
				args.entries = rf.log[rf.nextIndex[i]:]
			}

			reply := &AppendEntriesReply{}
			rf.sendAppendEntries(i, args, reply)

			//其他节点任期号大于当前节点任期号
			if reply.term > rf.currentTerm {
				exit <- struct{}{}
				return
			}

			//如果节点日志信息匹配上了
			if reply.success {
				rf.nextIndex[i] = len(rf.log) //指向下一个位置
				rf.persist()
				return
			}

			//当前位置没有日志
			if reply.xTerm == -1 {
				rf.nextIndex[i] = rf.nextIndex[i] - reply.term
				rf.persist()
				return
			}

			//节点日志信息没有匹配
			//返回的日志位置也不匹配
			if rf.log[reply.xIndex].Term != reply.term {
				//更新不匹配的日志的位置, 下次传输
				rf.nextIndex[i] = reply.xIndex
				rf.persist()
				return
			}

			//返回的日志位置匹配应该reply.success = true
		}(exit, i)
	}

	select {
	case <-exit:
		rf.stateConverter(leader, follower)
		return
	case <-rf.sm.heartBeatTimer.C:
		rf.stateConverter(leader, leader)
		return
	case <-rf.sm.killedMsg:
		return
	}
}

func (rf *Raft) listenAppendEntries() {
}
