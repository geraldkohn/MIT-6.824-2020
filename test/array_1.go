package main

import "fmt"

func main() {
	a := make([]int, 0)
	a = append(a, 1)
	for i := 0; i < len(a); i++ {
		a[i] = i * 3
	}

	for i := range a {
		fmt.Print(i)
	}

	fmt.Print("\n")
}