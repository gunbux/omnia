// Package currency parses conversion queries like "100 usd to eur" and resolves
// them against exchange rates fetched from open.er-api.com (cached on disk).
package currency

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/xdg"
)

const (
	ratesURL   = "https://open.er-api.com/v6/latest/USD"
	cacheTTL   = 12 * time.Hour
	fetchLimit = 8 * time.Second
)

// DefaultTargets are shown when a query has no explicit target currency.
var DefaultTargets = []string{"USD", "EUR", "GBP", "JPY", "SGD", "CNY"}

var symbols = map[string]string{
	"$": "USD", "€": "EUR", "£": "GBP", "¥": "JPY", "₹": "INR", "₩": "KRW",
	"₽": "RUB", "฿": "THB", "₫": "VND", "₱": "PHP", "₺": "TRY", "₪": "ILS",
	"₦": "NGN", "₴": "UAH", "zł": "PLN", "r$": "BRL", "s$": "SGD", "a$": "AUD",
	"c$": "CAD", "hk$": "HKD", "nz$": "NZD", "rm": "MYR", "kr": "SEK", "fr": "CHF",
}

// Names maps ISO codes to display names; it also acts as the set of known codes.
var Names = map[string]string{
	"AED": "UAE Dirham", "ARS": "Argentine Peso", "AUD": "Australian Dollar",
	"BDT": "Bangladeshi Taka", "BGN": "Bulgarian Lev", "BHD": "Bahraini Dinar",
	"BRL": "Brazilian Real", "CAD": "Canadian Dollar", "CHF": "Swiss Franc",
	"CLP": "Chilean Peso", "CNY": "Chinese Yuan", "COP": "Colombian Peso",
	"CZK": "Czech Koruna", "DKK": "Danish Krone", "EGP": "Egyptian Pound",
	"EUR": "Euro", "GBP": "British Pound", "HKD": "Hong Kong Dollar",
	"HUF": "Hungarian Forint", "IDR": "Indonesian Rupiah", "ILS": "Israeli Shekel",
	"INR": "Indian Rupee", "ISK": "Icelandic Króna", "JPY": "Japanese Yen",
	"KES": "Kenyan Shilling", "KRW": "South Korean Won", "KWD": "Kuwaiti Dinar",
	"LKR": "Sri Lankan Rupee", "MAD": "Moroccan Dirham", "MXN": "Mexican Peso",
	"MYR": "Malaysian Ringgit", "NGN": "Nigerian Naira", "NOK": "Norwegian Krone",
	"NZD": "New Zealand Dollar", "PEN": "Peruvian Sol", "PHP": "Philippine Peso",
	"PKR": "Pakistani Rupee", "PLN": "Polish Złoty", "QAR": "Qatari Riyal",
	"RON": "Romanian Leu", "RSD": "Serbian Dinar", "RUB": "Russian Ruble",
	"SAR": "Saudi Riyal", "SEK": "Swedish Krona", "SGD": "Singapore Dollar",
	"THB": "Thai Baht", "TRY": "Turkish Lira", "TWD": "New Taiwan Dollar",
	"UAH": "Ukrainian Hryvnia", "USD": "US Dollar", "VND": "Vietnamese Dong",
	"ZAR": "South African Rand",
}

// Query is a parsed conversion request.
type Query struct {
	Amount float64
	From   string
	To     string // empty means "show DefaultTargets"
}

var (
	symbolClass = `\$|€|£|¥|₹|₩|₽|฿|₫|₱|₺|₪|₦|₴|zł|r\$|s\$|a\$|c\$|hk\$|nz\$|rm|kr|fr`
	amount      = `(\d[\d,_]*(?:\.\d+)?|\.\d+)\s*([kmb])?`
	code        = `([a-zA-Z]{3})`
	sep         = `(?:\s*(?:to|in|into|as|->|→|=)\s*|\s+)`

	// 100 usd to eur | 100 usd eur | 100usd | 100 usd
	reAmountCode = regexp.MustCompile(`(?i)^` + amount + `\s*` + code + `(?:` + sep + `(?:` + code + `|(` + symbolClass + `)))?$`)
	// $100 to eur | €50 | $100 eur
	reSymbolAmount = regexp.MustCompile(`(?i)^(` + symbolClass + `)\s*` + amount + `(?:` + sep + `(?:` + code + `|(` + symbolClass + `)))?$`)
	// 100$ to eur | 100€
	reAmountSymbol = regexp.MustCompile(`(?i)^` + amount + `\s*(` + symbolClass + `)(?:` + sep + `(?:` + code + `|(` + symbolClass + `)))?$`)
	// usd to eur | usd eur (rate lookup, amount 1)
	reCodeCode = regexp.MustCompile(`(?i)^` + code + sep + code + `$`)
)

// Parse extracts a conversion query from free text. ok is false when the text
// is not a currency conversion (including unknown currency codes).
func Parse(input string) (q Query, ok bool) {
	s := strings.TrimSpace(input)
	if s == "" {
		return q, false
	}

	if m := reAmountCode.FindStringSubmatch(s); m != nil {
		return build(m[1], m[2], m[3], m[4], m[5])
	}
	if m := reSymbolAmount.FindStringSubmatch(s); m != nil {
		return build(m[2], m[3], m[1], m[4], m[5])
	}
	if m := reAmountSymbol.FindStringSubmatch(s); m != nil {
		return build(m[1], m[2], m[3], m[4], m[5])
	}
	if m := reCodeCode.FindStringSubmatch(s); m != nil {
		return build("1", "", m[1], m[2], "")
	}
	return q, false
}

func build(amountText, suffix, from, toCode, toSymbol string) (Query, bool) {
	amountText = strings.NewReplacer(",", "", "_", "").Replace(amountText)
	v, err := strconv.ParseFloat(amountText, 64)
	if err != nil {
		return Query{}, false
	}
	switch strings.ToLower(suffix) {
	case "k":
		v *= 1e3
	case "m":
		v *= 1e6
	case "b":
		v *= 1e9
	}

	fromCode, ok := normalize(from)
	if !ok {
		return Query{}, false
	}
	to := toCode
	if to == "" {
		to = toSymbol
	}
	var target string
	if to != "" {
		target, ok = normalize(to)
		if !ok {
			return Query{}, false
		}
	}
	return Query{Amount: v, From: fromCode, To: target}, true
}

func normalize(s string) (string, bool) {
	low := strings.ToLower(s)
	if c, ok := symbols[low]; ok {
		return c, true
	}
	up := strings.ToUpper(s)
	if _, ok := Names[up]; ok {
		return up, true
	}
	return "", false
}

// Rates holds USD-based exchange rates.
type Rates struct {
	FetchedAt time.Time          `json:"fetched_at"`
	Updated   string             `json:"updated"`
	Rates     map[string]float64 `json:"rates"`
}

// Convert returns amount expressed in the target currency.
func (r *Rates) Convert(amount float64, from, to string) (float64, error) {
	fromRate, ok := r.Rates[from]
	if !ok {
		return 0, fmt.Errorf("no rate for %s", from)
	}
	toRate, ok := r.Rates[to]
	if !ok {
		return 0, fmt.Errorf("no rate for %s", to)
	}
	return amount / fromRate * toRate, nil
}

// Stale reports whether the rates are older than the cache TTL.
func (r *Rates) Stale() bool {
	return time.Since(r.FetchedAt) > cacheTTL
}

func cachePath() string {
	return filepath.Join(xdg.CacheHome, "omnia", "rates.json")
}

// LoadCached returns rates from the on-disk cache, if any.
func LoadCached() (*Rates, error) {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return nil, err
	}
	var r Rates
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.Rates) == 0 {
		return nil, errors.New("empty cache")
	}
	return &r, nil
}

// Fetch downloads fresh rates and writes them to the cache.
func Fetch() (*Rates, error) {
	client := &http.Client{Timeout: fetchLimit}
	resp, err := client.Get(ratesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rates server returned %s", resp.Status)
	}

	var payload struct {
		Result  string             `json:"result"`
		Updated string             `json:"time_last_update_utc"`
		Rates   map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Result != "success" || len(payload.Rates) == 0 {
		return nil, errors.New("rates server returned no data")
	}

	r := &Rates{FetchedAt: time.Now(), Updated: payload.Updated, Rates: payload.Rates}
	if data, err := json.Marshal(r); err == nil {
		_ = os.MkdirAll(filepath.Dir(cachePath()), 0o755)
		_ = os.WriteFile(cachePath(), data, 0o644)
	}
	return r, nil
}

// Load returns cached rates when fresh, otherwise fetches. A stale cache is
// returned if fetching fails so the launcher still works offline.
func Load() (*Rates, error) {
	cached, cacheErr := LoadCached()
	if cacheErr == nil && !cached.Stale() {
		return cached, nil
	}
	fresh, err := Fetch()
	if err != nil {
		if cached != nil {
			return cached, nil
		}
		return nil, err
	}
	return fresh, nil
}

// FormatAmount renders a money amount with sensible precision.
func FormatAmount(v float64) string {
	switch {
	case v == 0:
		return "0"
	case v >= 1000:
		return addThousands(strconv.FormatFloat(v, 'f', 2, 64))
	case v >= 1:
		return strconv.FormatFloat(v, 'f', 2, 64)
	case v >= 0.01:
		return strconv.FormatFloat(v, 'f', 4, 64)
	default:
		return strconv.FormatFloat(v, 'g', 4, 64)
	}
}

func addThousands(s string) string {
	intPart, frac, _ := strings.Cut(s, ".")
	var b strings.Builder
	for i, ch := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(ch)
	}
	if frac != "" {
		b.WriteByte('.')
		b.WriteString(frac)
	}
	return b.String()
}
