package production

import (
	"testing"

	"github.com/google/uuid"
)

func TestAllocateFIFO(t *testing.T) {
	b1, b2 := uuid.New(), uuid.New()
	batches := []fifoBatch{
		{id: b1, qtyRemaining: 10, unitCost: 5},
		{id: b2, qtyRemaining: 20, unitCost: 6},
	}

	tests := []struct {
		name          string
		batches       []fifoBatch
		needed        float64
		wantCost      float64
		wantShortfall float64
		wantDeducts   int
	}{
		{"partial first batch", batches, 4, 4 * 5, 0, 1},
		{"exact first batch", batches, 10, 10 * 5, 0, 1},
		{"spill into second", batches, 15, 10*5 + 5*6, 0, 2},
		{"consume everything", batches, 30, 10*5 + 20*6, 0, 2},
		{"shortfall when over", batches, 35, 10*5 + 20*6, 5, 2},
		{"empty batches", nil, 5, 0, 5, 0},
		{"zero needed", batches, 0, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deductions, cost, shortfall := allocateFIFO(tt.batches, tt.needed)
			if !approx(cost, tt.wantCost) {
				t.Errorf("cost = %.4f, want %.4f", cost, tt.wantCost)
			}
			if !approx(shortfall, tt.wantShortfall) {
				t.Errorf("shortfall = %.4f, want %.4f", shortfall, tt.wantShortfall)
			}
			if len(deductions) != tt.wantDeducts {
				t.Errorf("deductions = %d, want %d", len(deductions), tt.wantDeducts)
			}
		})
	}
}

func TestAllocateFIFO_OrderAndAmounts(t *testing.T) {
	b1, b2 := uuid.New(), uuid.New()
	batches := []fifoBatch{
		{id: b1, qtyRemaining: 10, unitCost: 5},
		{id: b2, qtyRemaining: 20, unitCost: 6},
	}
	deductions, _, _ := allocateFIFO(batches, 15)

	if deductions[0].batchID != b1 || !approx(deductions[0].deduct, 10) {
		t.Errorf("first deduction = %+v, want batch b1 deduct 10", deductions[0])
	}
	if deductions[1].batchID != b2 || !approx(deductions[1].deduct, 5) {
		t.Errorf("second deduction = %+v, want batch b2 deduct 5", deductions[1])
	}
}

func TestAllocateFIFO_SkipsZeroQtyBatch(t *testing.T) {
	b1, b2 := uuid.New(), uuid.New()
	batches := []fifoBatch{
		{id: b1, qtyRemaining: 0, unitCost: 5},
		{id: b2, qtyRemaining: 8, unitCost: 6},
	}
	deductions, cost, shortfall := allocateFIFO(batches, 3)
	if len(deductions) != 1 || deductions[0].batchID != b2 {
		t.Errorf("expected single deduction from b2, got %+v", deductions)
	}
	if !approx(cost, 3*6) || shortfall != 0 {
		t.Errorf("cost=%.4f shortfall=%.4f, want 18 and 0", cost, shortfall)
	}
}

func approx(a, b float64) bool {
	d := a - b
	return d < 1e-9 && d > -1e-9
}
