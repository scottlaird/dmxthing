package main

import (
	"fmt"

	"github.com/google/gousb"
	"github.com/scottlaird/loupedeck"
	"github.com/scottlaird/udmx"

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

type AeonLightControl struct {
	Loupedeck *loupedeck.Loupedeck
	DMX *udmx.UDMXDevice
	DMXId uint16
	Position LoupedeckArea
	DialButtonId int
	X, Y int
	Brightness *loupedeck.WatchedInt
	ColorTemp *loupedeck.WatchedInt
	knob *loupedeck.IntKnob
	display *loupedeck.Display
}

func NewAeonLightControl(l *loupedeck.Loupedeck, dmx *udmx.UDMXDevice, dmxid uint16, position LoupedeckArea) *AeonLightControl {
	a := &AeonLightControl{
		Loupedeck: l,
		DMX: dmx,
		DMXId: dmxid,
		Position: position,
		Brightness: loupedeck.NewWatchedInt(0),
		ColorTemp: loupedeck.NewWatchedInt(128),
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
	_ = AeonColorTempButton(l, a.ColorTemp, dmx, int(dmxid+1), touch)

	a.Brightness.AddWatcher(func(i int) { a.DMX.Set(a.DMXId, uint16(i)) })
	a.Brightness.AddWatcher(func(i int) { a.Draw() })
	a.ColorTemp.AddWatcher(func(i int) { a.DMX.Set(a.DMXId+1, uint16(i)) })
	a.ColorTemp.AddWatcher(func(i int) { a.Draw() })

	return a
}

func (a *AeonLightControl) Draw() {
	im := image.NewRGBA(image.Rect(0, 0, 60, 90))
	bg := color.RGBA{0, 0, 0, 255}
	draw.Draw(im, im.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	fd := a.Loupedeck.FontDrawer()
	fd.Dst = im

	baseline := 55
	drawRightJustifiedStringAt(fd, strconv.Itoa(a.Brightness.Get()), 48, baseline)

	a.display.Draw(im, a.X, a.Y)
}

func drawRightJustifiedStringAt(fd font.Drawer, s string, x, y int) {
	bounds, _ := fd.BoundString(s)
	width := bounds.Max.X - bounds.Min.X
	x26 := fixed.I(x)
	y26 := fixed.I(y)

	fd.Dot = fixed.Point26_6{x26 - width, y26}
	fd.DrawString(s)
}

func main() {
	ctx := gousb.NewContext()
	defer ctx.Close()

	fmt.Printf("Connecting to DMX\n")

	udmx, err := udmx.NewUDMXDevice(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Trying loupedeck.\n")
	l, err := loupedeck.ConnectAuto()
	if err != nil {
		panic(err)
	}
	//defer l.Close()

	fmt.Printf("Connected.\n")

	go l.Listen()

	l.SetDisplays()

	light1 := NewAeonLightControl(l, udmx, 1, LeftTop)
	light2 := NewAeonLightControl(l, udmx, 3, LeftMiddle)
	light3 := NewAeonLightControl(l, udmx, 5, LeftBottom)
	
	light4 := loupedeck.NewWatchedInt(0)
	light4.AddWatcher(func(i int) { fmt.Printf("DMX 7->%d\n", i); udmx.Set(7, uint16(i)) })
	light5 := loupedeck.NewWatchedInt(0)
	light5.AddWatcher(func (i int) {
		v := i * 16
		l := v & 0xff
		h := v >> 8
		fmt.Printf("DMX 13+14->%d (%x %x)\n", i, h, l);
		udmx.Set(13, uint16(h)) 
		udmx.Set(14, uint16(l))
	})
	light6 := loupedeck.NewWatchedInt(0)
	light6.AddWatcher(func (i int) { fmt.Printf("DMX 11->%d\n", i); udmx.Set(11, uint16(i)) })
	//	fmt.Printf("Defined lights\n")

	right := l.GetDisplay("right")

	_ = l.NewTouchDial(right, light4, light5, light6, 0, 100) // Might actually be 255

	w4 := loupedeck.NewWatchedInt(0)
	w5 := loupedeck.NewWatchedInt(0)
	w6 := loupedeck.NewWatchedInt(0)

	_ = RainbowColorTempButton(l, w4, udmx, 8, loupedeck.Touch4)
	_ = RainbowColorTempButton(l, w5, udmx, 10, loupedeck.Touch8)
	_ = RainbowColorTempButton(l, w6, udmx, 15, loupedeck.Touch12)

	//	fmt.Printf("Defined colors\n")

	// Define the 'Circle' button (bottom left) to function as an "off" button.
	l.BindButton(loupedeck.Circle, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		l.SetButtonColor(loupedeck.Circle, color.RGBA{255, 0, 0, 255})
		light1.Brightness.Set(0)
		light2.Brightness.Set(0)
		light3.Brightness.Set(0)
		
		light4.Set(0)
		light5.Set(0)
		light6.Set(0)

		light1.ColorTemp.Set(128)
		light2.ColorTemp.Set(128)
		light3.ColorTemp.Set(128)
		w4.Set(1)
		w5.Set(1)
		w6.Set(1)
	})

	l.BindButton(loupedeck.Button1, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		l.SetButtonColor(loupedeck.Button1, color.RGBA{0, 255, 0, 255})
		light1.Brightness.Set(15)
		light2.Brightness.Set(3)
		light3.Brightness.Set(5)
		light4.Set(47)
		light5.Set(0)
		light6.Set(100)
		light1.ColorTemp.Set(128)
		light2.ColorTemp.Set(128)
		light3.ColorTemp.Set(128)
		w4.Set(172)
		w5.Set(172)
		w6.Set(172)
	})

	select {} // Wait forever
}

func AeonColorTempButton(l *loupedeck.Loupedeck, w *loupedeck.WatchedInt, udmx *udmx.UDMXDevice, dmxid int, button loupedeck.TouchButton) *loupedeck.MultiButton {
	fmt.Printf("Watchedint starts as %p=%#v\n", w, *w)
	ims := make([]image.Image, len(AeonColorTemps))
	for i, t := range AeonColorTemps {
		im, err := l.TextInBox(90, 90, t.Name, color.Black, t.Color)
		if err != nil {
			panic(err)
		}
		ims[i] = im
	}

	watchfunc := func(i int) { fmt.Printf("dmx%d -> %d\n", dmxid, i); udmx.Set(uint16(dmxid), uint16(i)) }
	w.AddWatcher(watchfunc)
	m := l.NewMultiButton(w, button, ims[0], AeonColorTemps[0].Value)
	for i := 1; i < len(ims); i++ {
		m.Add(ims[i], AeonColorTemps[i].Value)
	}
	//	fmt.Printf("Created new button: %#v\n", m)
	//	fmt.Printf("Watchedint is %#v\n", *w)

	return m
}


func RainbowColorTempButton(l *loupedeck.Loupedeck, w *loupedeck.WatchedInt, udmx *udmx.UDMXDevice, dmxid int, button loupedeck.TouchButton) *loupedeck.MultiButton {
	fmt.Printf("Watchedint starts as %p=%#v\n", w, *w)
	ims := make([]image.Image, len(RainbowColorTemps))
	for i, t := range RainbowColorTemps {
		im, err := l.TextInBox(90, 90, t.Name, color.Black, t.Color)
		if err != nil {
			panic(err)
		}
		ims[i] = im
	}

	watchfunc := func(i int) { fmt.Printf("dmx%d -> %d\n", dmxid, i); udmx.Set(uint16(dmxid), uint16(i)) }
	w.AddWatcher(watchfunc)
	m := l.NewMultiButton(w, button, ims[0], RainbowColorTemps[0].Value)
	for i := 1; i < len(ims); i++ {
		m.Add(ims[i], RainbowColorTemps[i].Value)
	}
	//	fmt.Printf("Created new button: %#v\n", m)
	//	fmt.Printf("Watchedint is %#v\n", *w)

	return m
}

type ColorTemp struct {
	Name  string
	Color *color.RGBA
	Value int
}

// Using RotoLight Aeon 2s, color temp 3150-6300K.  Using 5 values,
// assuming that there's roughly a linear relationship between Value
// and degrees K.  Using 3100-6300, that's a span of 3200K.  Let's
// have 5 points, spaced out every 800K.  So 3100, 3900, 4700, 5500,
// 6300.
//
// But those don't look far enough apart on the streamdeck, so let's
// push them apart further.  Let's leave 4700K the same but add an
// extra 800K to the RGB numbers.
var AeonColorTemps = []ColorTemp{
	{
		// Around 3100K, but use 1500K for visualization
		Name:  "3100K",
		Color: &color.RGBA{255, 109, 0, 255},
		Value: 1,
	},
	{
		// Around 3900K, but use 3100K for visualization
		Name:  "3900K",
		Color: &color.RGBA{255, 184, 114, 255},
		Value: 64,
	},
	{
		// Around 4700K
		Name:  "4700K",
		Color: &color.RGBA{255, 223, 194, 255},
		Value: 128,
	},
	{
		// Around 5500K, but use 7900K
		Name:  "5500K",
		Color: &color.RGBA{228, 234, 255, 255},
		Value: 192,
	},
	{
		// Around 6300K, but use 9500K
		Name:  "6300K",
		Color: &color.RGBA{208, 222, 255, 255},
		Value: 255,
	},
}

// Using Quasar Science Rainbow LEDs, color temp 2000-6000K.  Using 5
// values, assuming that there's roughly a linear relationship between
// Value and degrees K.  I'll assume (for now) that Quasar's DMX color
// temp is fairly accurate, with 0=2000K and 255=6000K, and a more or
// less linear relationship between the two.  I'm going to keep using
// the same settings at the Aeon, above, as much as possible.
var RainbowColorTemps = []ColorTemp{
	{
		// Around 2000K, but use 1500K for visualization
		Name:  "2000K",
		Color: &color.RGBA{255, 109, 0, 255},
		Value: 0,
	},
	{
		// Around 3900K, but use 3100K for visualization
		Name:  "3900K",
		Color: &color.RGBA{255, 184, 114, 255},
		Value: 121,
	},
	{
		// Around 4700K
		Name:  "4700K",
		Color: &color.RGBA{255, 223, 194, 255},
		Value: 172,
	},
	{
		// Around 5500K, but use 7900K
		Name:  "5500K",
		Color: &color.RGBA{228, 234, 255, 255},
		Value: 224,
	},
	{
		// Around 6000K, but use 9500K
		Name:  "6000K",
		Color: &color.RGBA{208, 222, 255, 255},
		Value: 255,
	},
}
