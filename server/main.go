package main

import (
	"fmt"
	"github.com/runningwild/jbot/game"
	"math/rand"
	"time"
)

func main() {
	host, err := game.MakeHost("127.0.0.1", 1231)
	if err != nil {
		fmt.Printf("Unable to start server: %v\n", err)
		return
	}
	fmt.Printf("Serving with %v ...\n", host)
	for {
		time.Sleep(time.Second)
		host.RLock()
		g := host.Game().(*game.Game)
		fmt.Printf("Status of g.Rebels: %v\n", g.Rebels)
		if g.Rebels != nil {
			fmt.Printf("Players are connected, load up a map.\n")
			level := game.Level{}
			level.Robots = make([]game.Robot, len(g.Rebels))
			level.Tiles = make([][]game.Tile, 5)
			for i := range level.Tiles {
				level.Tiles[i] = make([]game.Tile, 5)
				for j := range level.Tiles[i] {
					tt := game.TileEmpty
					if (i+j)%5 == 0 {
						tt = game.TilePit
					}
					level.Tiles[i][j] = game.Tile{
						Type: tt,
					}
				}
			}
			fmt.Printf("Unloacking...\n")
			host.RUnlock()
			fmt.Printf("Sending update...\n")
			host.MakeMajorUpdate(game.LoadLevel{Level: &level})
			fmt.Printf("Breaking...\n")
			break
		}
		host.RUnlock()
	}
	for {
		time.Sleep(time.Second)
		host.RLock()
		g := host.Game().(*game.Game)
		mode := g.Mode
		numRebels := len(g.Rebels)
		rebelsReady := g.RebelsReady
		host.RUnlock()
		if mode == game.ModeWaiting || (mode == game.ModeRun && rebelsReady) {
			cards := make([]game.Card, numRebels*7)
			for i := range cards {
				cards[i] = game.Card(rand.Intn(int(game.CardNumCards)))
			}
			host.MakeMajorUpdate(game.StartRound{Cards: cards})
			fmt.Printf("Sending cards: %v\n", cards)
		}
	}
}
