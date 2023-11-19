package dmxthing

import (
	"fmt"

	"github.com/scottlaird/udmx"
)


type AeosLight struct {
	dmx   *udmx.UDMXDevice
	dmxid uint16
}

func NewAeosLight(dmx *udmx.UDMXDevice, dmxid uint16) *AeosLight {
	a := &AeosLight{
		dmx:   dmx,
		dmxid: dmxid,
	}

	return a
}

func (a *AeosLight) SetBrightness(b int) {
	brightness := uint16(float64(b) * 2.55) // Input range is 0-100, output should be 0-255.
	a.dmx.Set(a.dmxid, brightness)
}

func (a *AeosLight) SetColorTemp(c int) {
	// Color points that I've been using (unverified, but seem reasonably close)
	//
	// 3150K = 1  (0 appears to be 'off'?
	// 3900K = 64
	// 4700K = 128
	// 5500K = 192
	// 6300K = 255

	// Map c=3150..6300 linearly onto 1..255.  This is *slightly*
	// different from the values above (4700k is 125.98 here, not
	// 128), but pretty close.
	v := uint16(((float32(c)-3150)*254/3150 + 1))
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

func (a *AeosLight) MinColorTemp() int {
	return 3150
}

func (a *AeosLight) MaxColorTemp() int {
	return 6300
}

