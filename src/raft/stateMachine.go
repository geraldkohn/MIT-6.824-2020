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
	state stateOfSM //状态机的状态

	electionTimer  *time.Timer //选举计时器
	heartBeatTimer *time.Timer //心跳计时器

	applyCh               chan ApplyMsg
	killChan              chan struct{} //节点被停止的信号
}

// 选举周期 300~450
func randomElectionTimeout() time.Duration {
	rand := rand.Intn(150) + 300
	return time.Duration(1000000 * rand) //ran毫秒
}

func (rf *Raft) leaderState() {

}

//stateChanger里面做状态初始化

func (rf *Raft) stateConverter(source stateOfSM, target stateOfSM) {

	//follower进程中选举超时
	if source == follower && target == candidate {
		//follower相关进程结束

		//candidate进程开始
		rf.votedFor = rf.me
		rf.currentTerm += 1
		rf.persist()
		rf.sm.electionTimer.Stop()
		rf.sm.electionTimer.Reset(randomElectionTimeout())
		rf.listenRequestVote()
		rf.sendRequestVoteToEachPeers()
		return
	}

	//Request Vote 投票给其他节点
	if source == follower && target == follower {
		rf.sm.electionTimer.Stop()
		rf.sm.electionTimer.Reset(randomElectionTimeout())
		rf.listenRequestVote()
		return
	}

	//Request Vote选举人接收到更大的选举周期, 或者半数投票反对(日志太旧)
	if source == candidate && target == follower {
		//candidate相关进程结束

		//follower进程开始
		rf.sm.state = follower
		rf.persist()
		rf.sm.electionTimer.Stop()
		rf.sm.electionTimer.Reset(randomElectionTimeout())
		rf.listenRequestVote()
		return
	}

	//sendRequestVote 超时
	if source == candidate && target == candidate {
		rf.sm.electionTimer.Stop()
		rf.sm.electionTimer.Reset(randomElectionTimeout())
		rf.sendRequestVoteToEachPeers()
	}

	//通过半数同意, candidate convert to leader
	if source == candidate && target == leader {
		//candidate相关进程结束

		//leader相关进程开始
		rf.sm.electionTimer.Stop()
		rf.sendAppendEntriesToEachPeers()
	}

	//Request Vote Leader接收到更大的选举周期
	if source == leader && target == follower {
		//leader相关进程结束

		//follower进程开始
		rf.sm.state = follower
		rf.persist()
		rf.sm.electionTimer.Stop()
		rf.sm.electionTimer.Reset(randomElectionTimeout())
		rf.listenRequestVote()
		return
	}
}

func (rf *Raft) init(target stateOfSM) {
	switch target {
	case follower:
		rf.sm.electionTimer.Reset(randomElectionTimeout())
	}
}
