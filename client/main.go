package main

import (
	"fmt"
	"github.com/runningwild/jbot/game"
	"math/rand"
	"os"
	"strings"
	"time"
)

func shuffle(v []int) {
	for i := range v {
		swap := rand.Intn(len(v)-i) + i
		v[i], v[swap] = v[swap], v[i]
	}
}

func main() {
	client, err := game.MakeClient("127.0.0.1", 1231)
	if err != nil {
		fmt.Printf("Unable to connect to server: %v\n", err)
		return
	}
	fmt.Printf("Connected with %v\n", client)

	client.MakeRequest(game.Join{Rebels: make([]*game.RebelPlayer, 2)})

	for {
		time.Sleep(time.Second)
		client.RLock()
		g := client.Game().(*game.Game)
		mode := g.Mode
		cards := g.Cards
		numRebels := len(g.Rebels)
		client.RUnlock()
		if mode == game.ModeProgram {
			var sp game.SubmitPrograms
			used := make(map[int]bool)
			for player := 0; player < numRebels; player++ {
				var program []int
				for program == nil {
					for i := range cards {
						if used[i] {
							continue
						}
						fmt.Printf("%d: %v\n", i, game.CardName(cards[i]))
					}
					var line []byte
					for {
						b := make([]byte, 1)
						_, err := os.Stdin.Read(b)
						if err != nil {
							panic(err)
						}
						if b[0] == '\n' {
							break
						}
						line = append(line, b[0])
					}
					fields := strings.Fields(string(line))
					for i := range fields {
						var n int
						_, err := fmt.Sscanf(fields[i], "%d", &n)
						if err != nil {
							break
						}
						if used[n] {
							break
						}
						used[n] = true
						program = append(program, n)
					}
					if len(program) != 5 {
						fmt.Printf("Invalid program.\n")
						program = nil
					}
					if program != nil {
						sp.Programs = append(sp.Programs, program)
					}
				}
			}
			fmt.Printf("Submitting Programs: %v\n", sp)
			client.MakeRequest(sp)
		} else if mode == game.ModeRun {
			fmt.Printf("Running programs...\n")
			time.Sleep(time.Second * 2)
			client.MakeRequest(game.Ready{Engineer: false})
		}
	}
}
