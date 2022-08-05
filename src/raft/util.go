package raft

import (
	"log"
)

// Debugging
const Debug = 1

// func DPrintf(format string, a ...interface{}) (n int, err error) {
// 	if Debug > 0 {
// 		log.Printf(format, a...)
// 	}
// 	return
// }

func (rf *Raft) DLog(format string, any ...interface{}) {
	if Debug == 0 {
		return
	}

	log.Printf(format, any...)
}
