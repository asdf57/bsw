package shared

import (
	"fmt"
	"strings"

	apimodels "github.com/asdf57/bsw/internal/models/api"
	currencylib "golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func FormatAmount(payment apimodels.Payment) string {
	currencyCode := strings.ToUpper(strings.TrimSpace(payment.Currency))
	if currencyCode == "" {
		return fmt.Sprintf("%.2f", payment.Amount)
	}

	unit, err := currencylib.ParseISO(currencyCode)
	if err != nil {
		return fmt.Sprintf("%.2f %s", payment.Amount, currencyCode)
	}

	p := message.NewPrinter(language.AmericanEnglish)
	return p.Sprintf("%v", currencylib.Symbol(unit.Amount(payment.Amount)))
}
