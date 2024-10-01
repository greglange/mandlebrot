package mandelbrot

import "testing"

type calculateTest struct {
	a, b            float64
	maxIteration    int
	inSet           bool
	escapeIteration int
}

var calculateTests = []calculateTest{
	calculateTest{0, 0, 1000, true, 1000},
	calculateTest{-1, 0, 1000, true, 1000},
	calculateTest{-1.1, 0, 1000, true, 1000},
	calculateTest{-1.3, 0, 1000, true, 1000},
	calculateTest{-1.38, 0, 1000, true, 1000},
	calculateTest{1, 0, 1000, false, 3},
	calculateTest{0, 1, 1000, true, 1000},
	calculateTest{0, 2, 1000, false, 2},
}

func TestCalculate(t *testing.T) {
	for _, test := range calculateTests {
		inSet, escapeIteration := Calculate(test.a, test.b, test.maxIteration)
		if inSet != test.inSet || escapeIteration != test.escapeIteration {
			t.Errorf("Unexpected return: (%t, %d) for: (%f, %f, %d)",
				inSet, escapeIteration, test.a, test.b, test.maxIteration)
		}
	}
}
