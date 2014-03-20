package main

import (
	"fmt"
	"github.com/runningwild/jbot/game"
)

func main() {
	client, err := game.MakeClient("127.0.0.1", 1231)
	if err != nil {
		fmt.Printf("Unable to connect to server: %v\n", err)
		return
	}
	fmt.Printf("Connected with %v\n", client)
}
