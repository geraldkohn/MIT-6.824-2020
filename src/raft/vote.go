package raft

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	term         int //candidate's term
	candidateId  int //candidate requesting vote
	lastLogIndex int //index of candidate's last log entry
	lastLogTerm  int //term of candidate's last log entry
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	term        int  //currentTerm. for candidate to update itself
	voteGranted bool //true means candidate received vote
}

//
// example RequestVote RPC handler.
//
// 1. reply false if term < currentTerm
// 2. if votedFor is null or candidateId, and candidate's log is at least as
// 	  up-to-date as receiver's log. grant vote.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	//接受rpc请求
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer func() {
		rf.DebugLog("%+v get request vote, args:%+v, reply:%+v", rf.sm.state, args, reply)
	}()

	//判断候选人的term和当前节点的term
	if args.term < rf.currentTerm {
		//任期落后, 拒绝投票
		reply.term = rf.currentTerm
		reply.voteGranted = false
		return
	} else if args.term == rf.currentTerm {
		switch rf.sm.state {
		case leader:
			//当前的角色是Leader, 拒绝投票
			reply.voteGranted = false
			reply.term = rf.currentTerm
			return

		case candidate:
			//当前的角色是candidate, 投自己的票
			reply.voteGranted = false
			reply.term = rf.currentTerm
			return

		case follower:
			if rf.votedFor == args.candidateId || rf.votedFor == -1 {
				//当前任期重复投票给同一个候选人, 或者没有投票, 判断是否选举限制
				lastLogTerm, lastLogIndex := rf.lastApplied()
				if args.lastLogTerm < lastLogTerm || args.lastLogTerm == lastLogTerm && args.lastLogIndex < lastLogIndex {
					//选举限制, 候选人日志落后于此节点. 拒绝选举
					reply.voteGranted = false
					reply.term = rf.currentTerm
					return
				} else {
					//选举不限制, 可以选举
					reply.voteGranted = true
					reply.term = rf.currentTerm
					//变换节点状态, follower --> follower
					rf.stateConverter(follower, follower)
					return
				}
			}

			if rf.votedFor != args.candidateId {
				//当前任期已经投票给其他候选人
				reply.voteGranted = false
				reply.term = rf.currentTerm
				return
			}
		}
	} else {
		//args.term > rf.currentTerm
		//任期落后. 更新任期号, 并且切换成跟随者状态. 并且重新储存需要持久化的数据
		defer rf.persist()
		rf.currentTerm = args.term
		rf.votedFor = -1

		switch rf.sm.state {
		case follower:
			lastLogTerm, lastLogIndex := rf.lastApplied()
			if args.lastLogTerm < lastLogTerm || args.lastLogTerm == lastLogTerm && args.lastLogIndex < lastLogIndex {
				//选举限制, 候选人日志落后于此节点. 拒绝选举
				reply.voteGranted = false
				reply.term = rf.currentTerm
				return
			} else {
				//选举不限制, 可以选举
				reply.voteGranted = true
				reply.term = rf.currentTerm
				//变换节点状态, follower --> follower
				rf.stateConverter(follower, follower)
				return
			}

		case candidate:
			//candidate状态转换成follower
			rf.stateConverter(candidate, follower)
			return

		case leader:
			//leader状态转换成follower
			rf.stateConverter(leader, follower)
			return
		}
	}
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	//候选人发送rpc请求
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendRequestVoteToEachPeers() {
	//响应RequestVote rpc节点个数
	respondPeer := 1
	//给候选人投票节点个数, 候选人自己给自己投票
	grantedPeer := 1
	//收到更高term, 需要退出函数
	exit := make(chan struct{})
	voteCh := make(chan bool, len(rf.peers))
	lastLogIndex, lastLogTerm := rf.lastApplied()
	args := &RequestVoteArgs{
		term:         rf.currentTerm,
		candidateId:  rf.me,
		lastLogIndex: lastLogIndex,
		lastLogTerm:  lastLogTerm,
	}
	//遍历每一个节点, 除了自己
	for i, _ := range rf.peers {
		if i == rf.me {
			continue
		}
		go func(ch chan bool, exit chan struct{}, i int) {
			reply := &RequestVoteReply{}
			rf.sendRequestVote(i, args, reply)
			ch <- reply.voteGranted
			//其他节点term比自身还要高, 需要退回到follower状态
			if reply.term > args.term {
				rf.stateConverter(candidate, follower)
				exit <- struct{}{}
			}
		}(voteCh, exit, i)
	}

	for {
		select {

		case <-exit:	//主进程退出
			close(voteCh)
			close(exit)
			return

		case <-rf.sm.electionTimer.C:	//选举超时
			close(voteCh)
			close(exit)
			rf.stateConverter(candidate, candidate)
			return

		case vote := <-voteCh:	//获得rpc调用结果
			respondPeer += 1
			if vote == true {
				grantedPeer += 1
			}
			//超过半数同意, candidate convert to leader
			if grantedPeer > len(rf.peers)/2 {
				rf.stateConverter(candidate, leader)
				return
			}
			//超过半数反对, candidate convert to follower
			if respondPeer-grantedPeer > len(rf.peers)/2 {
				rf.stateConverter(candidate, follower)
				return
			}
		}

	}

}

func (rf *Raft) listenRequestVote() {
}
