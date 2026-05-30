package shared

import (
	"fmt"
	"strings"

	apimodels "github.com/asdf57/bsw/internal/models/api"
	"github.com/shopspring/decimal"
)

func FormatAmount(payment apimodels.Payment) string {
	return FormatDecimalAmount(decimal.NewFromFloat(payment.Amount), payment.Currency)
}

func FormatDecimalAmount(amount decimal.Decimal, currency string) string {
	currencyCode := strings.ToUpper(strings.TrimSpace(currency))
	if currencyCode == "" {
		return amount.StringFixed(2)
	}

	symbol, ok := currencySymbols[currencyCode]
	if !ok {
		return fmt.Sprintf("%s %s", amount.StringFixed(2), currencyCode)
	}

	return symbol + formatDecimalWithCommas(amount, 2)
}

var currencySymbols = map[string]string{
	"AUD": "A$",
	"CAD": "C$",
	"CHF": "CHF ",
	"CNY": "¥",
	"EUR": "€",
	"GBP": "£",
	"JPY": "¥",
	"KRW": "₩",
	"USD": "$",
}

func formatDecimalWithCommas(amount decimal.Decimal, places int32) string {
	value := amount.StringFixed(places)
	sign := ""
	if strings.HasPrefix(value, "-") {
		sign = "-"
		value = strings.TrimPrefix(value, "-")
	}

	parts := strings.SplitN(value, ".", 2)
	whole := parts[0]
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + "," + whole[i:]
	}

	if len(parts) == 1 {
		return sign + whole
	}
	return sign + whole + "." + parts[1]
}
