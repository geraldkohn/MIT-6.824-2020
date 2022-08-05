package raft

type AppendEntriesArgs struct {
	Term         int //leader's term
	LeaderId     int //so follower can redirect clients
	PrevLogIndex int //index of log entry immediately preceding new ones
	PrevLogTerm  int //term of preLogIndex entry
	Entries      []logEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
	XTerm   int //这个是Follower中与Leader冲突的Log对应的任期号。在之前（7.1）有介绍Leader会在prevLogTerm中带上本地Log记录中，前一条Log的任期号。如果Follower在对应位置的任期号不匹配，它会拒绝Leader的AppendEntries消息，并将自己的任期号放在XTerm中。如果Follower在对应位置没有Log，那么这里会返回 -1
	XIndex  int //这个是Follower中，对应任期号为XTerm的第一条Log条目的槽位号
	XLen    int //如果Follower在对应位置没有Log，那么XTerm会返回-1，XLen表示空白的Log槽位数
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.DLog("serverId: %d; 收到AppendEntries rpc请求.", rf.me)
	if rf.sm.killed {
		return
	}

	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.DLog("get entries: %+v", args)

	//判断任期
	if rf.sm.state == candidate {
		if args.Term >= rf.currentTerm {
			rf.stateConverter(candidate, follower)
		}
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	//判断任期
	if rf.sm.state == leader {
		if args.Term >= rf.currentTerm {
			//需要状态转换
			rf.stateConverter(leader, follower)
		}
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	//当前节点term比leaderterm高
	if rf.currentTerm > args.Term {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	/** 快速恢复日志 */

	//当前位置没有日志
	//日志从下标为1开始, 但是日志最后的下标仍然为 len() - 1
	if len(rf.log)-1 < args.PrevLogIndex {
		reply.XTerm = -1
		reply.XLen = args.PrevLogIndex - len(rf.log)
		reply.Success = false
		reply.Term = rf.currentTerm
	}

	if rf.log[args.PrevLogIndex].Term == args.PrevLogTerm {
		//index同term同, 之前都相同
		reply.Term = rf.currentTerm
		reply.Success = true
		//更新节点日志
		rf.DLog("serverId: %d; 更新日志!", rf.me)
		rf.log = rf.log[:args.PrevLogIndex+1]
		rf.log = append(rf.log, args.Entries...)
		rf.persist()
		return
	} else {
		//index同Term不同, 需要继续向前找
		reply.XTerm = rf.log[args.PrevLogIndex].Term
		xIndex := 0
		var i int
		for i = args.PrevLogIndex; i > 1; i-- {
			if rf.log[i].Term != rf.log[args.PrevLogIndex].Term {
				xIndex = i + 1
				break
			}
		}
		if i == 0 {
			xIndex = 1
		}
		reply.XIndex = xIndex
		reply.Success = false
		reply.Term = rf.currentTerm
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
			rf.rLock(logEntriesMutex())
			args := &AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me,
				PrevLogIndex: rf.nextIndex[i] - 1,
				PrevLogTerm:  rf.log[rf.nextIndex[i]-1].Term,
				Entries:      nil,
				LeaderCommit: rf.commitIndex,
			}
			//上一次heartBeat到这一次heartBeat有命令
			if rf.nextIndex[i] < len(rf.log) {
				//防止数组越界
				args.Entries = rf.log[rf.nextIndex[i]:]
			}
			//发送请求之前的日志最后位置
			rf.matchIndex[i] = len(rf.log) - 1
			rf.rUnlock(logEntriesMutex())

			reply := &AppendEntriesReply{}
			rf.sendAppendEntries(i, args, reply)

			//其他节点任期号大于当前节点任期号
			if reply.Term > rf.currentTerm {
				exit <- struct{}{}
				return
			}

			//如果节点日志信息匹配上了
			if reply.Success {
				rf.DLog("serverId: %d; 更新该follower日志成功, 开始查找可以提交的日志.", rf.me)
				//更新leader可以提交的日志
				rf.updateCommittedIndex(i)
				//指向下一个位置, (这期间leader会收到很多请求, log在增长)
				rf.nextIndex[i] = len(rf.log)
				return
			}

			//当前位置没有日志
			if reply.XTerm == -1 {
				rf.nextIndex[i] = rf.nextIndex[i] - reply.Term
				return
			}

			//节点日志信息没有匹配
			//返回的日志位置也不匹配
			if rf.log[reply.XIndex].Term != reply.Term {
				//更新不匹配的日志的位置, 下次传输
				rf.nextIndex[i] = reply.XIndex
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
