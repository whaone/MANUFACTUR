package reports

import "testing"

func TestComputeMargin(t *testing.T) {
	tests := []struct {
		name       string
		sell, hpp  float64
		wantMargin float64
		wantPct    float64
	}{
		{"positive margin", 100, 60, 40, 40},
		{"zero margin", 50, 50, 0, 0},
		{"negative margin (loss)", 80, 100, -20, -25},
		{"zero sell price guards divide", 0, 30, -30, 0},
		{"negative sell price guards divide", -10, 5, -15, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			margin, pct := computeMargin(tt.sell, tt.hpp)
			if !approx(margin, tt.wantMargin) {
				t.Errorf("margin = %.4f, want %.4f", margin, tt.wantMargin)
			}
			if !approx(pct, tt.wantPct) {
				t.Errorf("marginPct = %.4f, want %.4f", pct, tt.wantPct)
			}
		})
	}
}

func approx(a, b float64) bool {
	d := a - b
	return d < 1e-9 && d > -1e-9
}
