package production

import "github.com/google/uuid"

// fifoBatch is a material batch available for consumption, oldest first.
type fifoBatch struct {
	id           uuid.UUID
	qtyRemaining float64
	unitCost     float64
}

// batchDeduction records how much to take from one batch and at what cost.
type batchDeduction struct {
	batchID  uuid.UUID
	deduct   float64
	unitCost float64
}

const fifoEpsilon = 1e-9

// allocateFIFO consumes `needed` quantity across batches in order, returning the
// per-batch deductions, accumulated cost, and any unfulfilled shortfall
// (shortfall > 0 means stock was insufficient). Pure: it neither touches the DB
// nor mutates its inputs, so it is unit-testable in isolation.
func allocateFIFO(batches []fifoBatch, needed float64) (deductions []batchDeduction, cost, shortfall float64) {
	for _, b := range batches {
		if needed <= fifoEpsilon {
			break
		}
		deduct := b.qtyRemaining
		if deduct > needed {
			deduct = needed
		}
		if deduct <= 0 {
			continue
		}
		needed -= deduct
		cost += deduct * b.unitCost
		deductions = append(deductions, batchDeduction{batchID: b.id, deduct: deduct, unitCost: b.unitCost})
	}
	if needed > fifoEpsilon {
		shortfall = needed
	}
	return deductions, cost, shortfall
}
