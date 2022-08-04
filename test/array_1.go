package main

import "fmt"

func main() {
	a := make([]int, 10)
	for i := 0; i < len(a); i++ {
		a[i] = i * 3
	}

	for i := range a {
		fmt.Print(i)
	}

	fmt.Print("\n")
}