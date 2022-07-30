package raft

import (
	"log"
)

// Debugging
const Debug = 0

// func DPrintf(format string, a ...interface{}) (n int, err error) {
// 	if Debug > 0 {
// 		log.Printf(format, a...)
// 	}
// 	return
// }

func (rf *Raft) DebugLog(format string, any ...interface{}) {
	if Debug == 0 {
		return
	}

	log.Printf(format, any...)
}
