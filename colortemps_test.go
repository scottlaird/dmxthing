package dmxthing

import (
	"slices"
	"testing"
)

type colorTest struct {
	min, max int
	want     []int
}

func TestBuildColorTemps(t *testing.T) {
	tests := []colorTest{
		{
			// Monochrome
			min:  5000,
			max:  5000,
			want: []int{5000},
		}, {
			// Exactly "standardColors"
			min:  3900,
			max:  5500,
			want: []int{3900, 4700, 5500},
		}, {
			// 190K lower at the bottom range, shouldn't add anything.
			min:  3710,
			max:  5500,
			want: []int{3900, 4700, 5500},
		}, {
			// 200K lower at the bottom range, should add 3700K to the output
			min:  3700,
			max:  5500,
			want: []int{3700, 3900, 4700, 5500},
		}, {
			// 190K higher at the top of the range, shouldn't add anything
			min:  3900,
			max:  5690,
			want: []int{3900, 4700, 5500},
		}, {
			// 200K higher at the top of the range, should add 5700K
			min:  3900,
			max:  5700,
			want: []int{3900, 4700, 5500, 5700},
		}, {
			// 800K extra on each end, should add one more color on each side.
			min:  3100,
			max:  6300,
			want: []int{3100, 3900, 4700, 5500, 6300},
		},
		// TODO: add ~3000K and verify that we get extra colors every 1000K or so.
	}

	for _, testdata := range tests {
		got := buildColorTemps(testdata.min, testdata.max)
		if slices.Compare(testdata.want, got) != 0 {
			t.Fatalf("buildColorTemps(%d,%d) got %v want %v", testdata.min, testdata.max, got, testdata.want)
		}
	}
}
