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

func mainLoop(client sgf.ClientEngine) {
	client.MakeRequest(game.Join{Rebels: make([]*game.RebelPlayer, 2)})
	ticker := time.Tick(time.Millisecond * 17)
	render.Queue(func() {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	})
	console := base.MakeConsole(wdx, wdy)
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
			programLoop(client, console)
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
func renderCardReticle(cardSize, cardX, cardY, x, y int) {
	cx := x + cardX*cardSize
	cy := y + cardY*cardSize
	base.EnableShader("reticle")
	base.SetUniformF("reticle", "edge", 0.05)
	base.SetUniformF("reticle", "size", 0.1)
	base.SetUniformF("reticle", "coverage", 1.0)
	gl.Color4ub(255, 255, 100, 200)
	texture.Render(float64(cx), float64(cy), float64(cardSize), float64(cardSize))
	base.EnableShader("")
}
func renderCards(cards []game.Card, cardSize, x, y, cols int) {
	gl.Enable(gl.TEXTURE_2D)
	cx := x - cardSize
	cy := y
	col := 0
	for i := range cards {
		if col == cols {
			col = 0
			cx = x - cardSize
			cy += cardSize
		}
		col++
		cx += cardSize
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
		tile.Render(float64(cx), float64(cy), float64(cardSize), float64(cardSize))
	}
}

func programLoop(client sgf.ClientEngine, console *base.Console) {
	ticker := time.Tick(time.Millisecond * 17)
	cols := 7
	sx, sy := 0, 0

	client.RLock()
	g := client.Game().(*game.Game)
	var cards []game.Card
	for _, card := range g.Cards {
		cards = append(cards, card)
	}
	client.RUnlock()

	for {
		<-ticker
		if gin.In().GetKey(gin.AnyEscape).FramePressCount() != 0 {
			return
		}
		sx -= gin.In().GetKey(gin.AnyKeyA).FramePressCount()
		sx += gin.In().GetKey(gin.AnyKeyD).FramePressCount()
		sy -= gin.In().GetKey(gin.AnyKeyW).FramePressCount()
		sy += gin.In().GetKey(gin.AnyKeyS).FramePressCount()
		if sx < 0 {
			sx = 0
		}
		if sx >= cols {
			sx = cols - 1
		}
		if sy < 0 {
			sy = 0
		}
		rows := (len(cards)-1)/cols + 1
		if sy >= rows {
			sy = rows - 1
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
			renderCards(cards, 64, 400, 400, cols)
			renderCardReticle(64, sx, sy, 400, 400)
		})
		render.Queue(func() {
			sys.SwapBuffers()
		})
		render.Purge()
	}
}

func main() {
	go server.Main()
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

	base.LoadAllDictionaries()

	client, err := game.MakeClient("127.0.0.1", 1231)
	if err != nil {
		base.Error().Printf("Unable to connect to server: %v\n", err)
		return
	}
	mainLoop(client)
}
