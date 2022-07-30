package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	t := time.NewTimer(2 * time.Second)
	ch := make(chan bool, 100)
	cnt := 0
	for i := 0; i < 100; i++ {
		go func(c chan bool) {
			time.Sleep(1 * time.Second)
			ch <- true
		}(ch)
	}

	for {
		select {
		case mess := <-ch:
			log.Println(mess)
			cnt++
		case <-t.C:
			fmt.Println("timeout")
			return
		}

		log.Println(cnt)
	}
}
