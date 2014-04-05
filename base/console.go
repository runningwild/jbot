package base

import (
	gl "github.com/chsc/gogl/gl21"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gui"
	"strings"
)

const maxLines = 30
const maxLineLength = 150
const lineHeight = 25

// A simple gui element that will display the last several lines of text from
// a log file (TODO: and also allow you to enter some basic commands).
type Console struct {
	lines   []string
	tail    Tailer
	dx, dy  int
	xscroll float64
	dict    *gui.Dictionary
	visible bool
}

func MakeConsole(dx, dy int) *Console {
	var c Console
	c.lines = make([]string, maxLines)
	c.tail = GetLogTailer()
	c.dict = GetDictionary("luxisr")
	c.dx = dx
	c.dy = dy
	return &c
}

func (c *Console) Draw(x, y, dx, dy int) {
	if !gin.In().GetKeyFlat(gin.EitherShift, gin.DeviceTypeAny, gin.DeviceIndexAny).IsDown() {
		return
	}
	c.tail.GetLines(c.lines)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Color4d(0.2, 0, 0.3, 0.8)
	gl.Disable(gl.TEXTURE_2D)
	gl.Begin(gl.QUADS)
	{
		glx := gl.Int(x)
		gly := gl.Int(y)
		glx2 := gl.Int(x + dx)
		gly2 := gl.Int(y + dy)
		gl.Vertex2i(glx, gly)
		gl.Vertex2i(glx, gly2)
		gl.Vertex2i(glx2, gly2)
		gl.Vertex2i(glx2, gly)
	}
	gl.End()
	gui.SetFontColor(1, 1, 1, 1)
	startY := float64(y + dy - len(c.lines)*lineHeight)
	for i, line := range c.lines {
		switch {
		case strings.HasPrefix(line, "LOG"):
			gui.SetFontColor(1, 1, 1, 1)
		case strings.HasPrefix(line, "WARN"):
			gui.SetFontColor(1, 1, 0, 1)
		case strings.HasPrefix(line, "ERROR"):
			gui.SetFontColor(1, 0, 0, 1)
		default:
			gui.SetFontColor(1, 1, 1, 0.7)
		}
		c.dict.RenderString(line, float64(x), startY+float64(i*lineHeight), 0, lineHeight, gui.Left)
	}
}
