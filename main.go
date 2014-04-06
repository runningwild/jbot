package main

import (
	"fmt"
	gl "github.com/chsc/gogl/gl21"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
	"github.com/runningwild/jbot/game"
	"github.com/runningwild/jbot/server"
	"github.com/runningwild/jbot/texture"
	"github.com/runningwild/sgf"
	"time"
	// "math"
	"github.com/runningwild/jbot/base"
	_ "image/png"
	"os"
	"path/filepath"
	"runtime"
)

var (
	sys      system.System
	datadir  string
	wdx, wdy int
	key_map  base.KeyMap
)

func init() {
	runtime.LockOSThread()
	sys = system.Make(gos.GetSystemInterface())

	datadir = filepath.Join(os.Args[0], "..", "..")
	base.SetDatadir(datadir)
	base.Log().Printf("Setting datadir: %s", datadir)
	wdx = 1000
	wdy = 800
	var key_binds base.KeyBinds
	base.LoadJson(filepath.Join(datadir, "key_binds.json"), &key_binds)
	fmt.Printf("Prething: %v\n", key_binds)
	key_map = key_binds.MakeKeyMap()
	base.SetDefaultKeyMap(key_map)
}

func mainLoop(client sgf.ClientEngine, controllers []gin.DeviceId, console *base.Console) {
	client.MakeRequest(game.Join{Rebels: make([]*game.RebelPlayer, 2)})
	ticker := time.Tick(time.Millisecond * 17)
	render.Queue(func() {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	})
	for {
		<-ticker
		if gin.In().GetKey(gin.AnyEscape).FramePressCount() != 0 {
			return
		}
		sys.Think()
		render.Queue(func() {
			gl.Clear(gl.COLOR_BUFFER_BIT)
			gl.Disable(gl.DEPTH_TEST)
			gui.SetFontColor(1, 1, 1, 1)
			gl.Disable(gl.TEXTURE_2D)
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()
			gl.Ortho(gl.Double(0), gl.Double(wdx), gl.Double(wdy), gl.Double(0), 1000, -1000)
			gl.ClearColor(0, 0, 0, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()
			base.GetDictionary("crackin").RenderString("Waiting on some nubs", float64(wdx)/2, 300, 0, 100, gui.Center)
		})

		client.RLock()
		g := client.Game().(*game.Game)
		mode := g.Mode
		client.RUnlock()
		if mode == game.ModeWaiting {
		} else if mode == game.ModeProgram {
			programLoop(client, controllers, console)
		} else if mode == game.ModeRun {

		}

		render.Queue(func() {
			sys.SwapBuffers()
		})
		render.Purge()
	}
}

func renderBoard(g *game.Game, x, y, dx, dy int) {
	tx := dx / len(g.Level.Tiles)
	ty := dy / len(g.Level.Tiles[0])
	gl.Enable(gl.TEXTURE_2D)
	gl.Color4ub(255, 255, 255, 255)
	for i := range g.Level.Tiles {
		for j := range g.Level.Tiles[i] {
			if g.Level.Tiles[i][j].Type == game.TileEmpty {
				tilePath := filepath.Join(base.GetDataDir(), fmt.Sprintf("textures/tile%d.png", (i+j)%4+1))
				tile := texture.LoadFromPath(tilePath)
				tile.Render(float64(x+i*tx), float64(y+j*ty), float64(tx), float64(ty))
			}
		}
	}
}
func renderCardReticle(selected bool, cardSize, cardX, cardY, x, y int) {
	base.Log().Printf("Rendering %d %d %d %d %d", cardSize, cardX, cardY, x, y)
	cx := x + cardX*cardSize
	cy := y + cardY*cardSize
	base.EnableShader("reticle")
	base.SetUniformF("reticle", "edge", 0.05)
	base.SetUniformF("reticle", "size", 0.1)
	if selected {
		base.SetUniformF("reticle", "coverage", 1.0)
	} else {
		base.SetUniformF("reticle", "coverage", 0.6)
	}
	texture.Render(float64(cx), float64(cy), float64(cardSize), float64(cardSize))
	base.EnableShader("")
}
func renderCards(cards []game.Card, cardSize, ulx, uly, cols int, cs *cardSelections) {
	gl.Enable(gl.TEXTURE_2D)
	x, y := -1, 0
	for i := range cards {
		if x == cols-1 {
			x = -1
			y++
		}
		x++
		cx := cardSize * x
		cy := cardSize * y
		var tile *texture.Data
		switch cards[i] {
		case game.CardForward1:
			tile = texture.LoadFromPath(filepath.Join(base.GetDataDir(), "textures/move1.png"))
		case game.CardForward2:
			tile = texture.LoadFromPath(filepath.Join(base.GetDataDir(), "textures/move2.png"))
		case game.CardForward3:
			tile = texture.LoadFromPath(filepath.Join(base.GetDataDir(), "textures/move3.png"))
		case game.CardReverse1:
			tile = texture.LoadFromPath(filepath.Join(base.GetDataDir(), "textures/reverse1.png"))
		case game.CardTurnLeft:
			tile = texture.LoadFromPath(filepath.Join(base.GetDataDir(), "textures/turn-left.png"))
		case game.CardTurnRight:
			tile = texture.LoadFromPath(filepath.Join(base.GetDataDir(), "textures/turn-right.png"))
		case game.CardUTurn:
			tile = texture.LoadFromPath(filepath.Join(base.GetDataDir(), "textures/u-turn.png"))
		}
		gl.Color4ub(255, 255, 255, 255)
		tile.Render(float64(cx+ulx), float64(cy+uly), float64(cardSize), float64(cardSize))
		if cs != nil && cs.used[i] != -1 {
			setColorForIndex(cs.used[i])
			renderCardReticle(true, cardSize, x, y, 400, 400)
		}
	}
}

type eventLogger struct{}

func (el *eventLogger) HandleEventGroup(group gin.EventGroup) {
	for _, event := range group.Events {
		base.Log().Printf("%v\n", event)
	}
}
func (el *eventLogger) Think() {}

type cardSelection struct {
	sx, sy  int
	cards   []game.Card
	indexes []int
}
type cardSelections struct {
	players []cardSelection
	cards   []game.Card
	used    []int // -1 means unused
	cols    int
}

const maxCardsPerProgram = 5

func (cs *cardSelections) HandleChoose(playerIndex int) {
	player := &cs.players[playerIndex]
	index := player.sx + player.sy*cs.cols
	if index < 0 || index >= len(cs.cards) {
		panic("FUCK YOU!!")
	}
	if len(player.cards) >= maxCardsPerProgram {
		return
	}
	if cs.used[index] == -1 {
		cs.used[index] = playerIndex
		player.cards = append(player.cards, cs.cards[index])
		player.indexes = append(player.indexes, index)
	}
}

func (cs *cardSelections) HandleDrop(playerIndex int) {
	player := &cs.players[playerIndex]
	index := player.sx + player.sy*cs.cols
	if index < 0 || index >= len(cs.cards) {
		panic("FUCK YOU!!")
	}
	if len(player.cards) > 0 {
		cs.used[player.indexes[len(player.indexes)-1]] = -1
		player.cards = player.cards[0 : len(player.cards)-1]
		player.indexes = player.indexes[0 : len(player.indexes)-1]
	}
}

func (cs *cardSelections) HandleMove(playerIndex int, dx, dy int) {
	if playerIndex < 0 || playerIndex >= len(cs.players) {
		panic("FUCK YOU")
	}
	player := &cs.players[playerIndex]
	player.sx += dx
	player.sy += dy
	if player.sx < 0 {
		player.sx = 0
	}
	if player.sx >= cs.cols {
		player.sx = cs.cols - 1
	}
	if player.sy < 0 {
		player.sy = 0
	}
	rows := (len(cs.cards)-1)/cs.cols + 1
	if player.sy >= rows {
		player.sy = rows - 1
	}
}

func setColorForIndex(index int) {
	switch index {
	case 0:
		gl.Color4ub(255, 0, 0, 200)
	case 1:
		gl.Color4ub(0, 255, 0, 200)
	case 2:
		gl.Color4ub(0, 0, 255, 200)
	default:
		gl.Color4ub(255, 0, 255, 200)
	}
}

func programLoop(client sgf.ClientEngine, controllers []gin.DeviceId, console *base.Console) {
	ticker := time.Tick(time.Millisecond * 17)
	var selections cardSelections
	selections.cols = 7
	selections.players = make([]cardSelection, len(controllers))
	client.RLock()
	g := client.Game().(*game.Game)
	for _, card := range g.Cards {
		selections.cards = append(selections.cards, card)
		selections.used = append(selections.used, -1)
	}
	client.RUnlock()
	for {
		<-ticker
		if gin.In().GetKey(gin.AnyEscape).FramePressCount() != 0 {
			return
		}
		for i, device := range controllers {
			up := gin.In().GetKeyFlat(gin.ControllerHatSwitchUp, device.Type, device.Index).FramePressCount()
			down := gin.In().GetKeyFlat(gin.ControllerHatSwitchDown, device.Type, device.Index).FramePressCount()
			left := gin.In().GetKeyFlat(gin.ControllerHatSwitchLeft, device.Type, device.Index).FramePressCount()
			right := gin.In().GetKeyFlat(gin.ControllerHatSwitchRight, device.Type, device.Index).FramePressCount()
			selections.HandleMove(i, right-left, down-up)
			drop := gin.In().GetKeyFlat(gin.ControllerButton0+1, device.Type, device.Index).FramePressCount() > 0
			choose := gin.In().GetKeyFlat(gin.ControllerButton0+2, device.Type, device.Index).FramePressCount() > 0
			if choose {
				selections.HandleChoose(i)
			}
			if drop {
				selections.HandleDrop(i)
			}
		}
		sys.Think()
		render.Queue(func() {
			defer console.Draw(0, 0, wdx, wdy)
			gl.Clear(gl.COLOR_BUFFER_BIT)
			gl.Disable(gl.DEPTH_TEST)
			gui.SetFontColor(1, 1, 1, 1)
			gl.Disable(gl.TEXTURE_2D)
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()
			gl.Ortho(gl.Double(0), gl.Double(wdx), gl.Double(wdy), gl.Double(0), 1000, -1000)
			gl.ClearColor(0, 0, 0, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()
			client.RLock()
			g := client.Game().(*game.Game)
			renderBoard(g, 10, 10, 400, 400)
			client.RUnlock()
			renderCards(selections.cards, 64, 400, 400, selections.cols, &selections)
			for i, player := range selections.players {
				setColorForIndex(i)
				renderCardReticle(false, 64, player.sx, player.sy, 400, 400)
				renderCards(player.cards, 64, 400, 300-100*i, selections.cols, nil)
			}
		})
		render.Queue(func() {
			sys.SwapBuffers()
		})
		render.Purge()
	}
}

type controllerTracker struct {
	ids map[gin.DeviceId]bool
}

func (ct *controllerTracker) HandleEventGroup(group gin.EventGroup) {
	if ct.ids == nil {
		ct.ids = make(map[gin.DeviceId]bool)
	}
	if group.Events[0].Key.Id().Device.Type == gin.DeviceTypeController {
		keyIndex := group.Events[0].Key.Id().Index
		deviceId := group.Events[0].Key.Id().Device
		switch keyIndex {
		case (gin.ControllerButton0 + 1):
			delete(ct.ids, deviceId)
		case (gin.ControllerButton0 + 2):
			ct.ids[deviceId] = group.Events[0].Key.IsDown()
		default:
			ct.ids[deviceId] = false
		}
	}
}
func (ct *controllerTracker) NumReady() int {
	n := 0
	for _, v := range ct.ids {
		if v {
			n++
		}
	}
	return n
}
func (ct *controllerTracker) Ready() bool {
	if len(ct.ids) == 0 {
		return false
	}
	for _, v := range ct.ids {
		if !v {
			return false
		}
	}
	return true
}
func (ct *controllerTracker) Think() {}

func getPlayers(console *base.Console) []gin.DeviceId {
	var ct controllerTracker
	gin.In().RegisterEventListener(&ct)
	defer gin.In().UnregisterEventListener(&ct)
	ticker := time.Tick(time.Millisecond * 17)
	start := time.Time{}
	readyDuration := time.Second * 2
	for start.IsZero() || time.Now().Sub(start) < readyDuration {
		<-ticker
		sys.Think()
		if ct.Ready() && start.IsZero() {
			start = time.Now()
		}
		if !ct.Ready() {
			start = time.Time{}
		}
		render.Queue(func() {
			defer console.Draw(0, 0, wdx, wdy)
			gl.Clear(gl.COLOR_BUFFER_BIT)
			gl.Disable(gl.DEPTH_TEST)
			gui.SetFontColor(1, 1, 1, 1)
			gl.Disable(gl.TEXTURE_2D)
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()
			gl.Ortho(gl.Double(0), gl.Double(wdx), gl.Double(wdy), gl.Double(0), 1000, -1000)
			gl.ClearColor(0, 0, 0, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()
			base.GetDictionary("crackin").RenderString(fmt.Sprintf("Num players: %d", len(ct.ids)), float64(wdx)/2, 300, 0, 100, gui.Center)
			base.GetDictionary("crackin").RenderString(fmt.Sprintf("Num ready: %d", ct.NumReady()), float64(wdx)/2, 400, 0, 100, gui.Center)
			if !start.IsZero() {
				base.GetDictionary("crackin").RenderString(fmt.Sprintf("Starting in %2.2f", (readyDuration-time.Now().Sub(start)).Seconds()), float64(wdx)/2, 500, 0, 100, gui.Center)
			}
		})
		render.Queue(func() {
			sys.SwapBuffers()
		})
		render.Purge()
	}
	var devices []gin.DeviceId
	for id := range ct.ids {
		devices = append(devices, id)
	}
	return devices
}

func main() {
	sys.Startup()
	err := gl.Init()
	if err != nil {
		base.Error().Fatalf("%v", err)
	}

	render.Init()
	render.Queue(func() {
		sys.CreateWindow(10, 10, wdx, wdy)
		sys.EnableVSync(true)
	})
	base.InitShaders()
	runtime.GOMAXPROCS(10)
	sys.Think()

	console := base.MakeConsole(wdx, wdy)
	controllers := getPlayers(console)
	base.Log().Printf("%v\n", controllers)
	go server.Main()

	base.LoadAllDictionaries()

	client, err := game.MakeClient("127.0.0.1", 1231)
	if err != nil {
		base.Error().Printf("Unable to connect to server: %v\n", err)
		return
	}
	mainLoop(client, controllers, console)
}
