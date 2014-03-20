package main

import (
	"fmt"
	"github.com/runningwild/jbot/game"
)

func main() {
	host, err := game.MakeHost("127.0.0.1", 1231)
	if err != nil {
		fmt.Printf("Unable to start server: %v\n", err)
		return
	}
	fmt.Printf("Serving with %v ...\n", host)
	select {}
}
