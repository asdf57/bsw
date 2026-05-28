package payment

import (
	"testing"
	"time"
)

func TestPaymentRangeFilterPresets(t *testing.T) {
	now := time.Date(2026, 5, 9, 15, 30, 0, 0, time.Local)

	tests := []struct {
		name      string
		value     string
		wantStart time.Time
		wantEnd   *time.Time
	}{
		{name: "today", value: PaymentRangeToday, wantStart: time.Date(2026, 5, 9, 0, 0, 0, 0, time.Local)},
		{
			name:      "yesterday",
			value:     PaymentRangeYesterday,
			wantStart: time.Date(2026, 5, 8, 0, 0, 0, 0, time.Local),
			wantEnd:   timePtr(time.Date(2026, 5, 9, 0, 0, 0, 0, time.Local)),
		},
		{name: "last 7 days", value: PaymentRangeLast7Days, wantStart: time.Date(2026, 5, 3, 0, 0, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := paymentRangeFilter(tt.value, now)
			if err != nil {
				t.Fatalf("paymentRangeFilter() error = %v", err)
			}
			if !got.start.Equal(tt.wantStart) {
				t.Fatalf("start = %s, want %s", got.start, tt.wantStart)
			}
			if (got.end == nil) != (tt.wantEnd == nil) {
				t.Fatalf("end = %v, want %v", got.end, tt.wantEnd)
			}
			if got.end != nil && !got.end.Equal(*tt.wantEnd) {
				t.Fatalf("end = %s, want %s", *got.end, *tt.wantEnd)
			}
		})
	}
}

func TestCustomPaymentRangeFilter(t *testing.T) {
	now := time.Date(2026, 5, 9, 15, 30, 0, 0, time.Local)

	tests := []struct {
		name      string
		input     string
		wantStart time.Time
	}{
		{name: "yyyy-mm-dd", input: "2026-05-01", wantStart: time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)},
		{name: "mm/dd/yyyy", input: "05/01/2026", wantStart: time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)},
		{name: "relative days", input: "10d", wantStart: time.Date(2026, 4, 30, 0, 0, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := customPaymentRangeFilter(tt.input, now)
			if err != nil {
				t.Fatalf("customPaymentRangeFilter() error = %v", err)
			}
			if !got.start.Equal(tt.wantStart) {
				t.Fatalf("start = %s, want %s", got.start, tt.wantStart)
			}
			if got.end != nil {
				t.Fatalf("end = %s, want nil", *got.end)
			}
		})
	}
}

func TestCustomPaymentRangeFilterRejectsInvalidInputs(t *testing.T) {
	now := time.Date(2026, 5, 9, 15, 30, 0, 0, time.Local)

	for _, input := range []string{"", "2026/05/01", "05-01-2026", "0d", "ten days"} {
		t.Run(input, func(t *testing.T) {
			if _, err := customPaymentRangeFilter(input, now); err == nil {
				t.Fatalf("customPaymentRangeFilter(%q) error = nil", input)
			}
		})
	}
}

func TestPaymentRangeFilterFromValues(t *testing.T) {
	now := time.Date(2026, 5, 9, 15, 30, 0, 0, time.Local)

	preset, err := paymentRangeFilterFromValues(PaymentRangeToday, "bad custom value", now)
	if err != nil {
		t.Fatalf("paymentRangeFilterFromValues() preset error = %v", err)
	}
	if !preset.start.Equal(time.Date(2026, 5, 9, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("preset start = %s", preset.start)
	}

	custom, err := paymentRangeFilterFromValues(PaymentRangeCustom, "10d", now)
	if err != nil {
		t.Fatalf("paymentRangeFilterFromValues() custom error = %v", err)
	}
	if !custom.start.Equal(time.Date(2026, 4, 30, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("custom start = %s", custom.start)
	}

	if _, err := paymentRangeFilterFromValues(PaymentRangeCustom, "", now); err == nil {
		t.Fatalf("paymentRangeFilterFromValues() custom empty error = nil")
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
