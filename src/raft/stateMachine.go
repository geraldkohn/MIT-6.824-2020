package raft

import (
	"math/rand"
	"time"
)

const (
	timeHeartBeat  = time.Millisecond * 150 // leader心跳周期
	timerpcTimeout = time.Millisecond * 100 // rpc超时时间
)

type stateMachine struct {
	state     stateOfSM     //状态机的状态
	killed    bool          //节点是否被终止
	killedMsg chan struct{} //节点终止信号

	electionTimer  *time.Timer //选举计时器
	heartBeatTimer *time.Timer //心跳计时器

	applyCh chan ApplyMsg
}

func (rf *Raft) reStart() {
	rf.sm.killed = false
	//初始化成follower
	rf.init(follower)
}

func (rf *Raft) firstStart() {
	//初始化所有资源
	rf.currentTerm = 1
	rf.votedFor = -1
	rf.log = make([]logEntry, 0)
	rf.persist()

	rf.sm.killed = false
	rf.sm.killedMsg = make(chan struct{}, 5)
	//初始化计时器, 防止uninitialized Timer panic
	rf.sm.electionTimer = time.NewTimer(10 * time.Second)
	rf.sm.electionTimer.Stop()
	rf.sm.heartBeatTimer = time.NewTimer(10 * time.Second)
	rf.sm.heartBeatTimer.Stop()

	//初始化为follower
	rf.init(follower)
}

func (rf *Raft) killStateMachine() {
	rf.sm.killed = true
	for i := 2; i > 0; i-- {
		rf.sm.killedMsg <- struct{}{}
	}
	rf.sm.electionTimer.Stop()
	rf.sm.heartBeatTimer.Stop()
}


// 选举周期 300~450
func randomElectionTimeout() time.Duration {
	rand := rand.Intn(150) + 300
	return time.Duration(1000000 * rand) //ran毫秒
}

//stateChanger里面做状态初始化

func (rf *Raft) stateConverter(source stateOfSM, target stateOfSM) {
	if rf.sm.killed {
		return
	}

	//follower进程中选举超时
	if source == follower && target == candidate {
		//follower相关进程结束
		rf.stop(follower)

		//candidate进程开始
		rf.init(candidate)
		return
	}

	//Request Vote 投票给其他节点
	if source == follower && target == follower {
		rf.sm.electionTimer.Stop()
		rf.sm.electionTimer = time.NewTimer(randomElectionTimeout())
		// rf.listenRequestVote()
		return
	}

	//Request Vote选举人接收到更大的选举周期, 或者半数投票反对(日志太旧)
	if source == candidate && target == follower {
		//candidate相关进程结束
		rf.stop(candidate)

		//follower进程开始
		rf.init(follower)
		return
	}

	//sendRequestVote 超时
	if source == candidate && target == candidate {
		rf.sm.electionTimer.Stop()
		rf.sm.electionTimer = time.NewTimer(randomElectionTimeout())
		rf.sendRequestVoteToEachPeers()
		return
	}

	//通过半数同意, candidate convert to leader
	if source == candidate && target == leader {
		//candidate相关进程结束
		rf.stop(candidate)

		//leader相关进程开始
		rf.init(leader)
		return
	}

	//Request Vote Leader接收到更大的选举周期
	if source == leader && target == follower {
		//leader相关进程结束
		rf.stop(leader)

		//follower进程开始
		rf.init(follower)
		return
	}

	//leader发送新一波心跳
	if source == leader && target == follower {
		rf.sm.heartBeatTimer.Stop()
		rf.sm.heartBeatTimer = time.NewTimer(timeHeartBeat)
		rf.sendAppendEntriesToEachPeers()
		return
	}

	//其他情况
	panic("stateConverter cannot handle input and output.")
}

func (rf *Raft) init(target stateOfSM) {
	switch target {
	case follower:
		rf.sm.state = follower
		rf.persist()
		rf.sm.electionTimer = time.NewTimer(randomElectionTimeout())
		// rf.listenRequestVote()
		// rf.listenAppendEntries()
	case candidate:
		rf.votedFor = rf.me
		rf.currentTerm += 1
		rf.sm.state = candidate
		rf.persist()
		rf.sm.electionTimer = time.NewTimer(randomElectionTimeout())
		// rf.listenRequestVote()
		// rf.listenAppendEntries()
		rf.sendRequestVoteToEachPeers()
	case leader:
		rf.sm.state = leader
		rf.nextIndex = make([]int, len(rf.peers))
		rf.matchIndex = make([]int, len(rf.peers))
		lastLogIndex, _ := rf.lastApplied()
		//初始化nextIndex数组
		for i := 0; i < len(rf.peers) && i != rf.me; i++ {
			rf.nextIndex[i] = lastLogIndex + 1
		}
		rf.persist()
		rf.sm.heartBeatTimer = time.NewTimer(timeHeartBeat)
		// rf.listenAppendEntries()
		// rf.listenRequestVote()
		rf.sendAppendEntriesToEachPeers()
	default:
		panic("state does not exist.")
	}
}

func (rf *Raft) stop(source stateOfSM) {
	switch source {
	case follower:
		rf.sm.electionTimer.Stop()
	case candidate:
		rf.sm.electionTimer.Stop()
	case leader:
		rf.sm.heartBeatTimer.Stop()
	default:
		panic("state does not exist.")
	}
}
