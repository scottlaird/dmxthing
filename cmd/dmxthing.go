package main

import (
	"fmt"

	"github.com/google/gousb"
	"github.com/scottlaird/dmxthing"
	"github.com/scottlaird/loupedeck"
	"github.com/scottlaird/udmx"

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

	l.SetDisplays()

	light1 := dmxthing.NewLightController(l, dmxthing.LeftTop, []dmxthing.Light{dmxthing.NewAeosLight(udmx, 1)})
	light2 := dmxthing.NewLightController(l, dmxthing.LeftMiddle, []dmxthing.Light{dmxthing.NewAeosLight(udmx, 3)})
	light3 := dmxthing.NewLightController(l, dmxthing.LeftBottom, []dmxthing.Light{dmxthing.NewAeosLight(udmx, 5)})

	light4 := dmxthing.NewLightController(l, dmxthing.RightTop, []dmxthing.Light{dmxthing.NewRainbowLightP5(udmx, 7)})
	light5 := dmxthing.NewLightController(l, dmxthing.RightMiddle, []dmxthing.Light{dmxthing.NewRainbowLightP12(udmx, 13)})
	light6 := dmxthing.NewLightController(l, dmxthing.RightBottom, []dmxthing.Light{dmxthing.NewRainbowLightP5(udmx, 11)})

	// Define the 'Circle' button (bottom left) to function as an "off" button.
	l.BindButton(loupedeck.Circle, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		l.SetButtonColor(loupedeck.Circle, color.RGBA{255, 0, 0, 255})
		light1.Reset()
		light2.Reset()
		light3.Reset()
		light4.Reset()
		light5.Reset()
		light6.Reset()
	})

	l.BindButton(loupedeck.Button1, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		l.SetButtonColor(loupedeck.Button1, color.RGBA{0, 255, 0, 255})
		light1.Brightness.Set(15)
		light2.Brightness.Set(3)
		light3.Brightness.Set(5)
		light4.Brightness.Set(47)
		light5.Brightness.Set(0)
		light6.Brightness.Set(100)
		light1.ColorTemp.Set(4700)
		light2.ColorTemp.Set(4700)
		light3.ColorTemp.Set(4700)
		light4.ColorTemp.Set(4700)
		light5.ColorTemp.Set(4700)
		light6.ColorTemp.Set(4700)
	})

	select {} // Wait forever
}
