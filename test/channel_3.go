package main

func main() {
	c := make(chan struct{}, 10)
	go func(c chan struct{}) {
		c <- struct{}{}
	}(c)
}
