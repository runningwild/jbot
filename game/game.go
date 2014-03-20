package game

import (
	"github.com/runningwild/sgf"
	"github.com/runningwild/sgf/types"
)

type Game struct {
	Mode     Mode
	Round    int
	Rebels   []*RebelPlayer
	Engineer *EngineerPlayer

	Level *Level

	local localData
}

type Mode int

const (
	ModeLobby Mode = iota
	ModePlaying
	ModeWrapup
)

type localData struct {
	// True iff this game is playing the rebels side
	rebels bool
}

type RebelPlayer struct{}
type EngineerPlayer struct{}
type Level struct{}

func MakeHost(addr string, port int) (sgf.HostEngine, error) {
	host, err := sgf.MakeHost(addr, port)
	if err != nil {
		return nil, err
	}
	host.RegisterGame(Game{})
	host.RegisterRequest(Join{})
	host.RegisterUpdate(Join{})
	host.Start(Game{})
	return host, nil
}

func MakeClient(addr string, port int) (sgf.ClientEngine, error) {
	client, err := sgf.MakeClient(addr, port)
	if err != nil {
		return nil, err
	}
	client.RegisterGame(Game{})
	client.RegisterRequest(Join{})
	client.RegisterUpdate(Join{})
	client.Start()
	return client, nil
}

type Join struct {
	Node     int
	Rebels   []*RebelPlayer
	Engineer *EngineerPlayer
}

func (j Join) ApplyRequest(node int, _game types.Game) []types.Update {
	game := _game.(*Game)
	if node != 0 {
		return nil
	}
	if (j.Rebels == nil) == (j.Engineer == nil) {
		return nil
	}
	if j.Rebels != nil && game.Rebels != nil {
		return nil
	}
	if j.Engineer != nil && game.Engineer != nil {
		return nil
	}
	j.Node = node
	game.local.rebels = (j.Rebels != nil)
	return []types.Update{j}
}
func (j Join) ApplyUpdate(node int, _game types.Game) {
	game := _game.(*Game)
	if node != 0 {
		return
	}
	if j.Rebels != nil {
		game.Rebels = j.Rebels
	} else {
		game.Engineer = j.Engineer
	}
}
