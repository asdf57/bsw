package shared

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFormatDecimalAmountUsesCurrencySymbol(t *testing.T) {
	got := FormatDecimalAmount(decimal.RequireFromString("1234.50"), "USD")
	if got != "$1,234.50" {
		t.Fatalf("FormatDecimalAmount() = %q, want %q", got, "$1,234.50")
	}
}

func TestFormatDecimalAmountFallsBackToCurrencyCode(t *testing.T) {
	got := FormatDecimalAmount(decimal.RequireFromString("12.30"), "XYZ")
	if got != "12.30 XYZ" {
		t.Fatalf("FormatDecimalAmount() = %q, want %q", got, "12.30 XYZ")
	}
}
