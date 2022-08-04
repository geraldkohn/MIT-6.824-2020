package main

import (
	"log"
	"time"
)

func main() {
	var timer *time.Timer
	// timer.Reset(time.Second)
	timer = time.NewTimer(time.Second)
	for {
		select{
		case <- timer.C:
			log.Println("timeout")
			// timer = time.NewTimer(time.Second)
			timer.Reset(time.Second)
		}
	}
}