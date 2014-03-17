package main

import (
	"fmt"
	gl "github.com/chsc/gogl/gl21"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
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

func mainLoop() {
	ticker := time.Tick(time.Millisecond * 17)
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
			base.GetDictionary("crackin").RenderString("J-Bots", 300, 300, 0, 100, gui.Left)
		})
		render.Queue(func() {
			sys.SwapBuffers()
		})
		render.Purge()
	}
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

	base.LoadAllDictionaries()

	mainLoop()
}
