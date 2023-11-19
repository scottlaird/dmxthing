package dmxthing

import (
	"fmt"

	"github.com/maruel/temperature"
	"github.com/scottlaird/loupedeck"

	"image"
	"image/color"

	"slices"
)

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
