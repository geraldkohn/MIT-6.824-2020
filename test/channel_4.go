package main

import (
	"fmt"
	"time"
)

func main() {
	c := make(chan bool, 1)

	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(1 * time.Second)
			c <- false
		}
		c <- true
	}()

	for {
		select {
		case b := <- c:
			if b {
				fmt.Printf("return\n")
				return
			} else {
				fmt.Printf("continue\n")
				continue
			}
		}
	}
}
