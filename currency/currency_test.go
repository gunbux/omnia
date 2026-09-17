package currency

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Query
	}{
		{"100 usd to eur", Query{100, "USD", "EUR"}},
		{"100 USD in SGD", Query{100, "USD", "SGD"}},
		{"100usd eur", Query{100, "USD", "EUR"}},
		{"100 usd", Query{100, "USD", ""}},
		{"$100 to eur", Query{100, "USD", "EUR"}},
		{"€50", Query{50, "EUR", ""}},
		{"£1,250.50 in usd", Query{1250.5, "GBP", "USD"}},
		{"100€ to $", Query{100, "EUR", "USD"}},
		{"2.5k sgd to myr", Query{2500, "SGD", "MYR"}},
		{"usd to jpy", Query{1, "USD", "JPY"}},
		{"1m jpy -> usd", Query{1e6, "JPY", "USD"}},
		{"S$20 in usd", Query{20, "SGD", "USD"}},
	}
	for _, c := range cases {
		got, ok := Parse(c.in)
		if !ok {
			t.Errorf("Parse(%q) not ok", c.in)
			continue
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestParseRejects(t *testing.T) {
	for _, in := range []string{"", "firefox", "100", "100 abc", "abc to def", "1+2", "100 usd to", "the cat", "100 myr to xyz"} {
		if q, ok := Parse(in); ok {
			t.Errorf("Parse(%q) = %+v, want not ok", in, q)
		}
	}
}

func TestConvert(t *testing.T) {
	r := &Rates{Rates: map[string]float64{"USD": 1, "EUR": 0.5, "SGD": 2}}
	got, err := r.Convert(100, "EUR", "SGD")
	if err != nil || got != 400 {
		t.Errorf("Convert = %v, %v; want 400", got, err)
	}
	if _, err := r.Convert(1, "USD", "XXX"); err == nil {
		t.Error("expected error for unknown code")
	}
}

func TestFormatAmount(t *testing.T) {
	cases := map[float64]string{
		1234567.891: "1,234,567.89",
		1234.5:      "1,234.50",
		99.999:      "100.00",
		0.5:         "0.5000",
		0.001234:    "0.001234",
	}
	for v, want := range cases {
		if got := FormatAmount(v); got != want {
			t.Errorf("FormatAmount(%v) = %q, want %q", v, got, want)
		}
	}
}
