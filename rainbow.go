package dmxthing

import (
	"fmt"

	"github.com/scottlaird/udmx"
)

// RainbowLight_P5 controls a Quasar Science Rainbow light bar (the
// original model, not Rainbow 2 or Double Rainbow), in DMX profile 5.
// In this mode, the light listens on two DMX addresses; the lower one
// is an 8-bit brightness and the upper one is an 8-bit color
// temperature.
type RainbowLight_P5 struct {
	dmx   *udmx.UDMXDevice
	dmxid uint16
}

// NewRainbowLight_P5 creates a new RainbowLight_P5 using a
// specific DMX controller and at a specific DMX address.
func NewRainbowLight_P5(dmx *udmx.UDMXDevice, dmxid uint16) *RainbowLight_P5 {
	a := &RainbowLight_P5{
		dmx:   dmx,
		dmxid: dmxid,
	}

	return a
}

// SetBrightness sets the brightness of the DMX light.
func (a *RainbowLight_P5) SetBrightness(b int) {
	brightness := uint16(float64(b) * 2.55) // Input range is 0-100, output should be 0-255.
	a.dmx.Set(a.dmxid, brightness)
}

// SetColorTemp sets the color temperature of the DMX light.
// The temperature should be specified in degrees K.
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

// RainbowLight_P12 controls a Quasar Science Rainbow light bar (the
// original model, not Rainbow 2 or Double Rainbow), in DMX profile
// 12.  In this mode, the light listens to a block of 12 DMX
// addresses:
//
// dmxid+0:  intensity (high 8 bits)
//
//	+1:  intensity (low 8 bits)
//	+2:  color temp (2000k to 6000k mapped onto 0..255)
//	+3:  plus/minus green
//	+4:  crossfade fraction color/rgb high
//	+5:  crossfade fraction color/rgb low
//	+6:  red
//	+7:  green
//	+8:  blue
//	+9:  FX
//	+10: FX rate
//	+11: FX size
//
// Right now, only intensity and color temp are used, but various
// special effect settings will be exposed in the future.
type RainbowLight_P12 struct {
	dmx   *udmx.UDMXDevice
	dmxid uint16
}

// NewRainbowLight_P12 creates a new RainbowLight_P12 using a
// specific DMX controller and at a specific DMX address.
func NewRainbowLight_P12(dmx *udmx.UDMXDevice, dmxid uint16) *RainbowLight_P12 {
	a := &RainbowLight_P12{
		dmx:   dmx,
		dmxid: dmxid,
	}

	return a
}

// SetBrightness sets the brightness of the DMX light.
func (a *RainbowLight_P12) SetBrightness(b int) {
	// In this mode, the light uses 16-bit brightness.  For my
	// use, I care more about the dimmest possible setting than I
	// do about fine control at the high/middle end, so I'm going
	// to map the 0..100 input onto the middle 8 bits for now.

	v := b * 16

	l := uint16(v & 0xff)
	h := uint16(v >> 8)

	a.dmx.Set(a.dmxid, h)
	a.dmx.Set(a.dmxid+1, l)
}

// SetColorTemp sets the color temperature of the DMX light.
// The temperature should be specified in degrees K.
func (a *RainbowLight_P12) SetColorTemp(c int) {
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

func (a *RainbowLight_P12) MinColorTemp() int {
	return 2000
}

func (a *RainbowLight_P12) MaxColorTemp() int {
	return 6000
}
