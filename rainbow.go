package dmxthing

import (
	"fmt"

	"github.com/scottlaird/udmx"
)

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
