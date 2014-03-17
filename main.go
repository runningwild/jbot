package main

import (
	"fmt"
	gl "github.com/chsc/gogl/gl21"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gos"
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
