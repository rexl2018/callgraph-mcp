package main

import (
	"fmt"
	"time"
)

func hello() {
	fmt.Println("Hello, World!")
}

func goodbye() {
	fmt.Println("Goodbye!")
}

func worker() {
	fmt.Println("worker start")
}

func main() {
	// initial call
	hello()

	// branch calls
	if time.Now().Unix()%2 == 0 {
		goodbye()
	} else {
		hello()
	}

	// start a goroutine
	go worker()
}