package currency

import (
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

const Default = "USD"

// Supported lists ISO 4217 codes for the price currency dropdown.
var Supported = []string{
	"USD", "EUR", "GBP", "CAD", "AUD", "JPY", "CHF", "CNY", "INR", "MXN",
	"BRL", "KRW", "NZD", "SEK", "NOK", "DKK", "PLN", "ZAR",
}

var supportedSet map[string]struct{}

func init() {
	supportedSet = make(map[string]struct{}, len(Supported))
	for _, code := range Supported {
		supportedSet[code] = struct{}{}
	}
}

func IsSupported(code string) bool {
	_, ok := supportedSet[Normalize(code)]
	return ok
}

func Normalize(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// DefaultFromPhone infers a currency code from an E.164 phone number.
func DefaultFromPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return Default
	}
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return Default
	}
	region := phonenumbers.GetRegionCodeForNumber(num)
	return DefaultFromRegion(region)
}

// DefaultFromRegion maps an ISO 3166-1 alpha-2 region to a supported currency.
func DefaultFromRegion(region string) string {
	region = strings.ToUpper(strings.TrimSpace(region))
	if region == "" {
		return Default
	}
	r, err := language.ParseRegion(region)
	if err != nil {
		return Default
	}
	unit, ok := currency.FromRegion(r)
	if !ok {
		return Default
	}
	code := unit.String()
	if IsSupported(code) {
		return code
	}
	return Default
}

// Format renders a price for display. Zero amount is always FREE.
func Format(amount int, code string) string {
	if amount == 0 {
		return "FREE"
	}
	code = Normalize(code)
	if code == "" {
		code = Default
	}
	unit, err := currency.ParseISO(code)
	if err != nil {
		return formatFallback(amount, code)
	}
	p := message.NewPrinter(language.English)
	sym := p.Sprint(currency.Symbol(unit))
	num := p.Sprint(number.Decimal(int64(amount), number.Scale(0)))
	return sym + " " + num
}

func formatFallback(amount int, code string) string {
	return strconv.Itoa(amount) + " " + code
}
