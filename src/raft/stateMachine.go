package raft

type stateMachine struct {
	//Persistent state on all servers
	currentTerm int //lastest term server has seen
	votedFor    int //candidateId that received vote in current term(or null if none)
	log         []LogEntry

	//
}
