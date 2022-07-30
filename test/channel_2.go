package main

import (
	"fmt"
	"time"
)

func main() {
	exit := make(chan struct{})
	go func(exit chan struct{}) {
		exit <- struct{}{}
		fmt.Println("sub progress return")
		return
	}(exit)
	
	select {
	case <-exit:
		return
	}
	time.Sleep(1 * time.Second)
	fmt.Println("main progress return")
}
