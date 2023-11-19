package main

import (
	"fmt"

	"github.com/google/gousb"
	"github.com/maruel/temperature"
	"github.com/scottlaird/loupedeck"
	"github.com/scottlaird/udmx"
	"github.com/scottlaird/dmxthing"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"

	"slices"
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

// Sets brightness to 0 and color to min
func (a *LightController) Reset() {
	a.Brightness.Set(0)
	a.ColorTemp.Set(a.minColorTemp)
}

type RainbowLight_P5 struct {
	dmx   *udmx.UDMXDevice
	dmxid uint16
}

func NewRainbowLight_P5(dmx *udmx.UDMXDevice, dmxid uint16) *RainbowLight_P5 {
	a := &RainbowLight_P5{
		dmx:   dmx,
		dmxid: dmxid,
	}

	return a
}

func (a *RainbowLight_P5) SetBrightness(b int) {
	brightness := uint16(float64(b) * 2.55) // Input range is 0-100, output should be 0-255.
	a.dmx.Set(a.dmxid, brightness)
}

func (a *RainbowLight_P5) SetColorTemp(c int) {
	// Map c=2000..6000 linearly onto 1..255.
	v := uint16(((float32(c)-2000)*254/4000 + 1))
	fmt.Printf("Setting color temp to %d for %dK\n", v, c)
	if v > 255 {
		// Entertainingly, sending weird enough values to UDMX
		// can crash the controller, so we're better off
		// panicking here than crashing it and requireing a
		// the controller to be power-cycled.
		panic("Color temp out of range!")
	}
	a.dmx.Set(a.dmxid, v)
}

func (a *RainbowLight_P5) MinColorTemp() int {
	return 2000
}

func (a *RainbowLight_P5) MaxColorTemp() int {
	return 6000
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

	light1 := NewLightController(l, LeftTop, []Light{dmxthing.NewAeosLight(udmx, 1)})
	light2 := NewLightController(l, LeftMiddle, []Light{dmxthing.NewAeosLight(udmx, 3)})
	light3 := NewLightController(l, LeftBottom, []Light{dmxthing.NewAeosLight(udmx, 5)})

	light4 := NewLightController(l, RightTop, []Light{NewRainbowLight_P5(udmx, 7)})
	//light5 := NewLightController(l, RightBottom, []Light{NewRainbowLight_P12(udmx, 13)})
	light6 := NewLightController(l, RightBottom, []Light{NewRainbowLight_P5(udmx, 11)})
	/*
		light5.AddWatcher(func(i int) {
			v := i * 16
			l := v & 0xff
			h := v >> 8
			fmt.Printf("DMX 13+14->%d (%x %x)\n", i, h, l)
			udmx.Set(13, uint16(h))
			udmx.Set(14, uint16(l))
		})
	*/
	//	fmt.Printf("Defined colors\n")

	// Define the 'Circle' button (bottom left) to function as an "off" button.
	l.BindButton(loupedeck.Circle, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		l.SetButtonColor(loupedeck.Circle, color.RGBA{255, 0, 0, 255})
		light1.Reset()
		light2.Reset()
		light3.Reset()
		light4.Reset()
		//light5.Reset()
		light6.Reset()
	})

	l.BindButton(loupedeck.Button1, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		l.SetButtonColor(loupedeck.Button1, color.RGBA{0, 255, 0, 255})
		light1.Brightness.Set(15)
		light2.Brightness.Set(3)
		light3.Brightness.Set(5)
		light4.Brightness.Set(47)
		//light5.Brightness.Set(0)
		light6.Brightness.Set(100)
		light1.ColorTemp.Set(4700)
		light2.ColorTemp.Set(4700)
		light3.ColorTemp.Set(4700)
		light4.ColorTemp.Set(4700)
		//light5.ColorTemp.Set(4700)
		light6.ColorTemp.Set(4700)
	})

	select {} // Wait forever
}

func buildColorTemps(min, max int) []int {
	// So we want to build a few buttons in the range from `min`K
	// to `max`K, including 3900K, 4700K, and 5500K.  This needs
	// to be dynamic, as I have lights with 3 different color temp
	// ranges, and I don't want to manage this manually.  Also,
	// I'd like to be able to group dissimilar lights into a
	// single controller and have the right thing happen.
	//
	// At a minimum, we probably want [min, 3900, 4700, 5500, and max].
	//
	// But we don't really want gaps of <200K, or over 1000K.  So,
	// if min is 3800, then we should either exclude 3800 or 3900.
	// I could argue that either is reasonable.  If min is 1200,
	// then we probably want to include something like 1200, 2000,
	// 3000, 3900,...  Arguably, those should be more like 1200,
	// 2100, 3000, 3900.
	//
	// If min > 3900 or max < 5500, then we need to exclude those.

	// First, let's handle the monochrome light case (min==max):
	if min == max {
		return []int{min}
	}

	colors := []int{}

	standardColors := []int{3900, 4700, 5500}
	minStandard := slices.Min(standardColors) // Oooh, ahh, Go 1.21 generics FTW.
	maxStandard := slices.Max(standardColors)

	if min < minStandard {
		minRange := minStandard - min

		if minRange < 200 {
			// omit 'min', because it's too close to
			// minStandard.  We'd rather have predictable
			// color temps than exercise the full range of
			// the light.
			colors = append(colors, minStandard)
		} else {
			colors = append(colors, min)
			colors = append(colors, minStandard)

			// TODO: add additional colors here when minRange is large.
		}
	} else {
		// Just add 'min'.
		colors = append(colors, min)
	}

	// Include additional standard colors, as long as they're
	// within min..max (and aren't minStandard or maxStandard)
	for _, c := range standardColors {
		if c > minStandard && c > min && c < max && c < maxStandard {
			colors = append(colors, c)
		}
	}

	// Should be the same as the min/minStandard block, above.
	if max > maxStandard {
		maxRange := max - maxStandard

		if maxRange < 200 {
			colors = append(colors, maxStandard)
		} else {
			colors = append(colors, max)
			colors = append(colors, maxStandard)
			// TODO: add additional colors as needed
		}
	} else {
		// just add max
		colors = append(colors, max)
	}

	slices.Sort(colors)
	return colors
}

func colorTempToRGB(temp int) color.Color {
	// Stretch the color out a bit to make it more obvious on the display.
	fakeTemp := 2*(temp-4700) + 4700

	// The lookup table only goes down to 1000K, and asking for
	// negative temps is Right Out due to unsigned integers and/or
	// physics.  So clamp to 1000K at the bottom.
	if fakeTemp < 1000 {
		fakeTemp = 1000
	}

	r, g, b := temperature.ToRGB(uint16(fakeTemp))
	return color.RGBA{r, g, b, 0xff}
}

func ColorTempButton(l *loupedeck.Loupedeck, w *loupedeck.WatchedInt, min, max int, button loupedeck.TouchButton) *loupedeck.MultiButton {
	colorTemps := buildColorTemps(min, max)
	fmt.Printf("***** COLOR TEMPS ARE: %v\n", colorTemps)

	ims := make([]image.Image, len(colorTemps))
	for i, t := range colorTemps {
		im, err := l.TextInBox(90, 90, fmt.Sprintf("%dK", t), color.Black, colorTempToRGB(t))
		if err != nil {
			panic(err)
		}
		ims[i] = im
	}

	fmt.Printf("* adding default button with temp=%d\n", colorTemps[0])
	m := l.NewMultiButton(w, button, ims[0], colorTemps[0])
	for i := 1; i < len(ims); i++ {
		fmt.Printf("* adding button with temp=%d\n", colorTemps[i])
		m.Add(ims[i], colorTemps[i])
	}

	return m
}
