package main

import (
	"fmt"
	"time"
)

func main() {
	t := time.NewTimer(1 * time.Second)
	t.Stop()
	t.Stop()
	fmt.Println("no panic")
}