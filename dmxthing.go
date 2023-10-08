package main

import (
	"fmt"

	"github.com/google/gousb"
	"github.com/scottlaird/loupedeck"
	"github.com/scottlaird/udmx"

	"image"
	"image/color"
)

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

	light1 := loupedeck.NewWatchedInt(0)
	light1.AddWatcher(func(i int) { fmt.Printf("DMX 1->%d\n", i); udmx.Set(1, uint16(i)) })
	light2 := loupedeck.NewWatchedInt(0)
	light2.AddWatcher(func(i int) { fmt.Printf("DMX 3->%d\n", i); udmx.Set(3, uint16(i)) })
	light3 := loupedeck.NewWatchedInt(0)
	light3.AddWatcher(func(i int) { fmt.Printf("DMX 5->%d\n", i); udmx.Set(5, uint16(i)) })
	light4 := loupedeck.NewWatchedInt(0)
	light4.AddWatcher(func(i int) { fmt.Printf("DMX 7->%d\n", i); udmx.Set(7, uint16(i)) })
	light5 := loupedeck.NewWatchedInt(0)
	//	light5.AddWatcher(func (i int) { fmt.Printf("DMX 9->%d\n", i) })
	light6 := loupedeck.NewWatchedInt(0)
	//	light6.AddWatcher(func (i int) { fmt.Printf("DMX 11->%d\n", i) })

	//	fmt.Printf("Defined lights\n")

	_ = l.NewTouchDial(loupedeck.DisplayLeft, light1, light2, light3, 0, 100)
	//	fmt.Printf("Defined touchdial 1\n")

	_ = l.NewTouchDial(loupedeck.DisplayRight, light4, light5, light6, 0, 100) // Might actually be 255

	//	fmt.Printf("Defined touchdials\n")

	w1 := loupedeck.NewWatchedInt(0)
	w2 := loupedeck.NewWatchedInt(0)
	w3 := loupedeck.NewWatchedInt(0)
	w4 := loupedeck.NewWatchedInt(0)
	w5 := loupedeck.NewWatchedInt(0)
	w6 := loupedeck.NewWatchedInt(0)
	_ = AeonColorTempButton(l, w1, udmx, 2, loupedeck.Touch1)
	_ = AeonColorTempButton(l, w2, udmx, 4, loupedeck.Touch5)
	_ = AeonColorTempButton(l, w3, udmx, 6, loupedeck.Touch9)
	_ = AeonColorTempButton(l, w4, udmx, 8, loupedeck.Touch4)
	_ = AeonColorTempButton(l, w5, udmx, 10, loupedeck.Touch8)
	_ = AeonColorTempButton(l, w6, udmx, 12, loupedeck.Touch12)

	//	fmt.Printf("Defined colors\n")

	// Define the 'Circle' button (bottom left) to function as an "off" button.
	l.BindButton(loupedeck.Circle, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		l.SetButtonColor(loupedeck.Circle, color.RGBA{255, 0, 0, 255})
		light1.Set(0)
		light2.Set(0)
		light3.Set(0)
		light4.Set(0)
		light5.Set(0)
		light6.Set(0)
		w1.Set(1)
		w2.Set(1)
		w3.Set(1)
		w4.Set(1)
		w5.Set(1)
		w6.Set(1)
	})

	l.BindButton(loupedeck.Button1, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		l.SetButtonColor(loupedeck.Button1, color.RGBA{0, 255, 0, 255})
		light1.Set(15)
		light2.Set(3)
		light3.Set(5)
		light4.Set(47)
		light5.Set(0)
		light6.Set(0)
		w1.Set(128)
		w2.Set(128)
		w3.Set(128)
		w4.Set(128)
		w5.Set(64)
		w6.Set(64)
	})

	select {} // Wait forever
}

func AeonColorTempButton(l *loupedeck.Loupedeck, w *loupedeck.WatchedInt, udmx *udmx.UDMXDevice, dmxid int, button loupedeck.TouchButton) *loupedeck.MultiButton {
	fmt.Printf("Watchedint starts as %p=%#v\n", w, *w)
	ims := make([]image.Image, len(ColorTemps))
	for i, t := range ColorTemps {
		im, err := l.TextInBox(90, 90, t.Name, color.Black, t.Color)
		if err != nil {
			panic(err)
		}
		ims[i] = im
	}

	watchfunc := func(i int) { fmt.Printf("dmx%d -> %d\n", dmxid, i); udmx.Set(uint16(dmxid), uint16(i)) }
	w.AddWatcher(watchfunc)
	m := l.NewMultiButton(w, button, ims[0], ColorTemps[0].Value)
	for i := 1; i < len(ims); i++ {
		m.Add(ims[i], ColorTemps[i].Value)
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
var ColorTemps = []ColorTemp{
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
