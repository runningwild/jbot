package game

import (
	"fmt"
	"github.com/runningwild/sgf"
	"github.com/runningwild/sgf/types"
)

type Game struct {
	Mode     Mode
	Round    int
	Rebels   []*RebelPlayer
	Engineer *EngineerPlayer

	Level    *Level
	Cards    []Card
	Programs [][]int

	RebelsReady   bool
	EngineerReady bool

	local localData
}

type Mode int

const (
	ModeLobby Mode = iota
	ModeWaiting
	ModeProgram
	ModeRun
	ModeWrapup
)

type localData struct {
	// True iff this game is playing the rebels side
	rebels bool
}

type RebelPlayer struct{}
type EngineerPlayer struct{}
type Level struct {
	Tiles  [][]Tile
	Robots []Robot
}
type Tile struct {
	Type TileType
	Dir  Direction
}

type TileType int

const (
	TileEmpty TileType = iota
	TileSingleConveyor
	TileDoubleConveyor
	TilePit
)

type Direction int

const (
	DirNone Direction = iota
	DirUp
	DirRight
	DirDown
	DirLeft
)

func (d Direction) TurnLeft() Direction {
	switch d {
	case DirUp:
		return DirLeft
	case DirLeft:
		return DirDown
	case DirDown:
		return DirRight
	case DirRight:
		return DirUp
	}
	panic(fmt.Sprintf("%v can't turn left.", d))
}

func (d Direction) TurnRight() Direction {
	switch d {
	case DirLeft:
		return DirUp
	case DirDown:
		return DirLeft
	case DirRight:
		return DirDown
	case DirUp:
		return DirRight
	}
	panic(fmt.Sprintf("%v can't turn right.", d))
}

type Robot struct {
	X, Y  int
	Dir   Direction
	Rebel bool
	Index int
}

type Card int

const (
	CardForward1 Card = iota
	CardForward2
	CardForward3
	CardReverse1
	CardTurnLeft
	CardTurnRight
	CardUTurn
	CardNumCards
)

func CardName(c Card) string {
	switch c {
	case CardForward1:
		return "Forward 1"
	case CardForward2:
		return "Forward 2"
	case CardForward3:
		return "Forward 3"
	case CardReverse1:
		return "Reverse 1"
	case CardTurnLeft:
		return "Turn Left"
	case CardTurnRight:
		return "Turn Right"
	case CardUTurn:
		return "U-Turn"
	default:
		panic(fmt.Sprintf("Unknown card: %v\n", c))
	}
}

func MakeHost(addr string, port int) (sgf.HostEngine, error) {
	host, err := sgf.MakeHost(addr, port)
	if err != nil {
		return nil, err
	}
	host.RegisterGame(&Game{})
	host.RegisterRequest(Join{})
	host.RegisterUpdate(Join{})
	host.RegisterRequest(SubmitPrograms{})
	host.RegisterUpdate(SubmitPrograms{})
	host.RegisterRequest(Ready{})
	host.RegisterUpdate(Ready{})
	host.RegisterUpdate(LoadLevel{})
	host.RegisterUpdate(StartRound{})
	host.Start(&Game{})
	return host, nil
}

func MakeClient(addr string, port int) (sgf.ClientEngine, error) {
	client, err := sgf.MakeClient(addr, port)
	if err != nil {
		return nil, err
	}
	client.RegisterGame(&Game{})
	client.RegisterRequest(Join{})
	client.RegisterUpdate(Join{})
	client.RegisterRequest(SubmitPrograms{})
	client.RegisterUpdate(SubmitPrograms{})
	client.RegisterRequest(Ready{})
	client.RegisterUpdate(Ready{})
	client.RegisterUpdate(LoadLevel{})
	client.RegisterUpdate(StartRound{})
	client.Start()
	return client, nil
}

type Join struct {
	Node     int
	Rebels   []*RebelPlayer
	Engineer *EngineerPlayer
}

func (j Join) ApplyRequest(node int, _game types.Game) []types.Update {
	fmt.Printf("Request join: %v\n", j)
	game := _game.(*Game)
	if (j.Rebels == nil) == (j.Engineer == nil) {
		fmt.Printf("Return on 1\n")
		return nil
	}
	if j.Rebels != nil && game.Rebels != nil {
		fmt.Printf("Return on 2\n")
		return nil
	}
	if j.Engineer != nil && game.Engineer != nil {
		fmt.Printf("Return on 3\n")
		return nil
	}
	j.Node = node
	game.local.rebels = (j.Rebels != nil)
	return []types.Update{j}
}
func (j Join) ApplyUpdate(node int, _game types.Game) {
	fmt.Printf("Update join on node %d: %v\n", node, j)
	game := _game.(*Game)
	if node != 0 {
		return
	}
	if j.Rebels != nil {
		fmt.Printf("Setting rebels to %v\n", j.Rebels)
		game.Rebels = j.Rebels
	} else {
		fmt.Printf("Setting Engineer to %v\n", j.Engineer)
		game.Engineer = j.Engineer
	}
}

type LoadLevel struct {
	Level *Level
}

func (ll LoadLevel) ApplyUpdate(node int, _game types.Game) {
	game := _game.(*Game)
	if node != 0 && node != -2 {
		fmt.Printf("Not doing LoadLevel, node is %d\n", node)
		return
	}
	fmt.Printf("Doing LoadLevel, node is %d\n", node)
	game.Level = ll.Level
	game.Mode = ModeWaiting
}

type StartRound struct {
	Cards []Card
}

func (sr StartRound) ApplyUpdate(node int, _game types.Game) {
	game := _game.(*Game)
	game.Cards = sr.Cards
	game.Mode = ModeProgram
	game.EngineerReady = false
	game.RebelsReady = false
}

type SubmitPrograms struct {
	// Program [i][j] is the i'th player's index into game.Cards
	Programs [][]int
}

func (sp SubmitPrograms) ApplyRequest(node int, _game types.Game) []types.Update {
	// TODO: verify that this data is sensible
	return []types.Update{sp}
}
func (j SubmitPrograms) ApplyUpdate(node int, _game types.Game) {
	game := _game.(*Game)
	game.Mode = ModeRun
	game.Programs = j.Programs
}

type Ready struct {
	Engineer bool
}

func (r Ready) ApplyRequest(node int, _game types.Game) []types.Update {
	// TODO: verify that this Ready makes sense
	return []types.Update{r}
}
func (j Ready) ApplyUpdate(node int, _game types.Game) {
	game := _game.(*Game)
	if j.Engineer {
		game.EngineerReady = true
	} else {
		game.RebelsReady = true
	}
}
