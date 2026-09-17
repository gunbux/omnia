package calc

import (
	"math"
	"testing"
)

func TestEval(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"1+2", 3},
		{"2*3+4", 10},
		{"2*(3+4)", 14},
		{"10/4", 2.5},
		{"2^10", 1024},
		{"2**3", 8},
		{"2^3^2", 512},
		{"-3^2", -9},
		{"5!", 120},
		{"50%", 0.5},
		{"200*10%", 20},
		{"10 % 3", 1},
		{"10 mod 3", 1},
		{"sqrt(16)", 4},
		{"2pi", 2 * math.Pi},
		{"2(3+4)", 14},
		{"min(3, 7)", 3},
		{"0xff", 255},
		{"0b101", 5},
		{"1e3", 1000},
		{"2e", 2 * math.E},
		{"1_000_000 / 4", 250000},
		{"2k * 3", 6000},
		{"1.5m", 1.5e6},
		{"3 × 4 ÷ 2", 6},
		{"abs(-2.5)", 2.5},
		{"log(1000)", 3},
		{"round(2.5)", 3},
	}
	for _, c := range cases {
		got, err := Eval(c.in)
		if err != nil {
			t.Errorf("Eval(%q) error: %v", c.in, err)
			continue
		}
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Eval(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestEvalErrors(t *testing.T) {
	for _, in := range []string{"", "1/0", "2+", "(1+2", "foo(1)", "abc", "1 2 3"} {
		if _, err := Eval(in); err == nil {
			t.Errorf("Eval(%q) expected error", in)
		}
	}
}

func TestLooksLikeExpression(t *testing.T) {
	yes := []string{"1+1", "2*pi", "sqrt(2)", "5!", "100/3", "(1+2)*3"}
	no := []string{"firefox", "42", "pi", "e", "code", "chrome-beta", "a+b", ""}
	for _, s := range yes {
		if !LooksLikeExpression(s) {
			t.Errorf("LooksLikeExpression(%q) = false, want true", s)
		}
	}
	for _, s := range no {
		if LooksLikeExpression(s) {
			t.Errorf("LooksLikeExpression(%q) = true, want false", s)
		}
	}
}

func TestFormat(t *testing.T) {
	cases := map[float64]string{
		3:                "3",
		2.5:              "2.5",
		1.0 / 3.0:        "0.333333333333",
		1e20:             "1e+20",
		-0.1 + 0.3 - 0.2: "0",
	}
	for v, want := range cases {
		if got := Format(v); got != want {
			t.Errorf("Format(%v) = %q, want %q", v, got, want)
		}
	}
}
