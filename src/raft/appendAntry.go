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
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {

}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntriesToEachPeers() {
	
}
