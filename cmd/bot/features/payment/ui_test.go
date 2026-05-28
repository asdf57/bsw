package payment

import (
	"testing"

	apimodels "github.com/asdf57/bsw/internal/models/api"
)

func TestPaymentEmbedBatches(t *testing.T) {
	payments := make([]apimodels.PaymentResponse, 27)
	for i := range payments {
		payments[i].ID = uint(i + 1)
	}

	batches := paymentEmbedBatches(payments, 10)
	if len(batches) != 3 {
		t.Fatalf("len(batches) = %d, want 3", len(batches))
	}

	wantSizes := []int{10, 10, 7}
	for i, wantSize := range wantSizes {
		if len(batches[i]) != wantSize {
			t.Fatalf("len(batches[%d]) = %d, want %d", i, len(batches[i]), wantSize)
		}
	}
}
