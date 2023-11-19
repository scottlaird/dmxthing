package dmxthing

import (
	"github.com/scottlaird/loupedeck"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"

	"strconv"
)

type LoupedeckArea int

const (
	LeftTop LoupedeckArea = iota
	LeftMiddle
	LeftBottom
	RightTop
	RightMiddle
	RightBottom
)

type Light interface {
	SetBrightness(int)
	SetColorTemp(int)
	MinColorTemp() int
	MaxColorTemp() int
}

type LightController struct {
	Loupedeck    *loupedeck.Loupedeck
	Position     LoupedeckArea
	DialButtonId int
	X, Y         int
	Brightness   *loupedeck.WatchedInt
	ColorTemp    *loupedeck.WatchedInt
	Lights       []Light
	knob         *loupedeck.IntKnob
	display      *loupedeck.Display
	minColorTemp int
}

func NewLightController(l *loupedeck.Loupedeck, position LoupedeckArea, lights []Light) *LightController {
	a := &LightController{
		Loupedeck:  l,
		Position:   position,
		Brightness: loupedeck.NewWatchedInt(0),
		ColorTemp:  loupedeck.NewWatchedInt(128),
		Lights:     lights,
	}

	var knob loupedeck.Knob
	var touch loupedeck.TouchButton

	switch position {
	case LeftTop:
		a.X, a.Y = 0, 0
		a.display = l.GetDisplay("left")
		knob = loupedeck.Knob1
		touch = loupedeck.Touch1
	case LeftMiddle:
		a.X, a.Y = 0, 90
		a.display = l.GetDisplay("left")
		knob = loupedeck.Knob2
		touch = loupedeck.Touch5
	case LeftBottom:
		a.X, a.Y = 0, 180
		a.display = l.GetDisplay("left")
		knob = loupedeck.Knob3
		touch = loupedeck.Touch9
	case RightTop:
		a.X, a.Y = 0, 0
		a.display = l.GetDisplay("right")
		knob = loupedeck.Knob4
		touch = loupedeck.Touch4
	case RightMiddle:
		a.X, a.Y = 0, 90
		a.display = l.GetDisplay("right")
		knob = loupedeck.Knob5
		touch = loupedeck.Touch8
	case RightBottom:
		a.X, a.Y = 0, 180
		a.display = l.GetDisplay("right")
		knob = loupedeck.Knob6
		touch = loupedeck.Touch12
	default:
		panic("Unknown position!")
	}

	a.knob = l.IntKnob(knob, 0, 100, a.Brightness)

	// Find minimum/maximum color temps for all lights in set.
	min := lights[0].MinColorTemp()
	max := lights[0].MaxColorTemp()
	for _, l := range lights {
		lMin := l.MinColorTemp()
		lMax := l.MaxColorTemp()
		if lMin < min {
			min = lMin
		}
		if lMax > max {
			max = lMax
		}
	}
	a.minColorTemp = min
	_ = ColorTempButton(l, a.ColorTemp, min, max, touch)

	a.Brightness.AddWatcher(func(i int) {
		for _, l := range a.Lights {
			l.SetBrightness(i)
		}
	})
	a.Brightness.AddWatcher(func(i int) { a.Draw() })
	a.ColorTemp.AddWatcher(func(i int) {
		for _, l := range a.Lights {
			l.SetColorTemp(i)
		}
	})
	a.ColorTemp.AddWatcher(func(i int) { a.Draw() })

	return a
}

// Draw updates the Loupedeck display.
func (a *LightController) Draw() {
	im := image.NewRGBA(image.Rect(0, 0, 60, 90))
	bg := color.RGBA{0, 0, 0, 255}
	draw.Draw(im, im.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	fd := a.Loupedeck.FontDrawer()
	fd.Dst = im

	baseline := 55
	drawRightJustifiedStringAt(fd, strconv.Itoa(a.Brightness.Get()), 48, baseline)

	a.display.Draw(im, a.X, a.Y)
}

// Reset sets the brightness to 0 and color to the minimum supported by the light.
func (a *LightController) Reset() {
	a.Brightness.Set(0)
	a.ColorTemp.Set(a.minColorTemp)
}

func drawRightJustifiedStringAt(fd font.Drawer, s string, x, y int) {
	bounds, _ := fd.BoundString(s)
	width := bounds.Max.X - bounds.Min.X
	x26 := fixed.I(x)
	y26 := fixed.I(y)

	fd.Dot = fixed.Point26_6{x26 - width, y26}
	fd.DrawString(s)
}
