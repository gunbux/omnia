// Package calc implements a small arithmetic expression evaluator for the launcher's calculator mode.
package calc

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

var ErrEmpty = errors.New("empty expression")

var constants = map[string]float64{
	"pi":  math.Pi,
	"e":   math.E,
	"tau": 2 * math.Pi,
	"phi": math.Phi,
}

var unaryFuncs = map[string]func(float64) float64{
	"sqrt":  math.Sqrt,
	"cbrt":  math.Cbrt,
	"abs":   math.Abs,
	"floor": math.Floor,
	"ceil":  math.Ceil,
	"round": math.Round,
	"trunc": math.Trunc,
	"ln":    math.Log,
	"log":   math.Log10,
	"log2":  math.Log2,
	"exp":   math.Exp,
	"sin":   math.Sin,
	"cos":   math.Cos,
	"tan":   math.Tan,
	"asin":  math.Asin,
	"acos":  math.Acos,
	"atan":  math.Atan,
	"sinh":  math.Sinh,
	"cosh":  math.Cosh,
	"tanh":  math.Tanh,
	"deg":   func(x float64) float64 { return x * 180 / math.Pi },
	"rad":   func(x float64) float64 { return x * math.Pi / 180 },
}

var binaryFuncs = map[string]func(float64, float64) float64{
	"min":   math.Min,
	"max":   math.Max,
	"pow":   math.Pow,
	"hypot": math.Hypot,
	"atan2": math.Atan2,
	"mod":   math.Mod,
}

// LooksLikeExpression reports whether the input is plausibly meant as a
// calculation rather than an app name: it must parse and contain an operator
// or function call, so bare words and bare numbers are left alone.
func LooksLikeExpression(input string) bool {
	s := strings.TrimSpace(input)
	if s == "" {
		return false
	}
	if !strings.ContainsAny(s, "+-*/^%!()×÷") && !containsFunc(s) {
		return false
	}
	if !strings.ContainsAny(s, "0123456789") && !containsConst(s) {
		return false
	}
	_, err := Eval(s)
	return err == nil
}

func containsFunc(s string) bool {
	low := strings.ToLower(s)
	for name := range unaryFuncs {
		if strings.Contains(low, name+"(") {
			return true
		}
	}
	for name := range binaryFuncs {
		if strings.Contains(low, name+"(") {
			return true
		}
	}
	return false
}

func containsConst(s string) bool {
	for _, word := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r)
	}) {
		if _, ok := constants[word]; ok {
			return true
		}
	}
	return false
}

// Eval parses and evaluates an arithmetic expression.
func Eval(input string) (float64, error) {
	p := &parser{src: []rune(strings.TrimSpace(input))}
	if len(p.src) == 0 {
		return 0, ErrEmpty
	}
	v, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	p.skipSpace()
	if p.pos < len(p.src) {
		return 0, fmt.Errorf("unexpected %q", string(p.src[p.pos]))
	}
	if math.IsNaN(v) {
		return 0, errors.New("not a number")
	}
	if math.IsInf(v, 0) {
		return 0, errors.New("infinite")
	}
	return v, nil
}

// Format renders a result compactly, trimming float noise.
func Format(v float64) string {
	if v == math.Trunc(v) && math.Abs(v) < 1e15 {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	s := strconv.FormatFloat(v, 'g', 12, 64)
	if strings.Contains(s, "e") {
		return s
	}
	// 'g' with precision may leave trailing zeros after the decimal point.
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}

// Alternates returns extra representations for integer results (hex, binary, octal).
func Alternates(v float64) []string {
	if v != math.Trunc(v) || math.Abs(v) >= 1<<53 || v < 0 {
		return nil
	}
	n := int64(v)
	if n < 2 {
		return nil
	}
	return []string{
		"0x" + strconv.FormatInt(n, 16),
		"0b" + strconv.FormatInt(n, 2),
		"0o" + strconv.FormatInt(n, 8),
	}
}

type parser struct {
	src []rune
	pos int
}

func (p *parser) skipSpace() {
	for p.pos < len(p.src) && unicode.IsSpace(p.src[p.pos]) {
		p.pos++
	}
}

func (p *parser) peek() rune {
	p.skipSpace()
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

// expr := term (('+' | '-') term)*
func (p *parser) parseExpr() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		switch p.peek() {
		case '+':
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left += right
		case '-':
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left -= right
		default:
			return left, nil
		}
	}
}

// term := unary (('*' | '/' | '%' | implicit) unary)*
func (p *parser) parseTerm() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	for {
		c := p.peek()
		switch {
		case c == '*' || c == '×':
			p.pos++
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			left *= right
		case c == '/' || c == '÷':
			p.pos++
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, errors.New("division by zero")
			}
			left /= right
		case c == '%':
			// '%' followed by an operand is modulo; otherwise it is a percent postfix.
			save := p.pos
			p.pos++
			if p.startsOperand() {
				right, err := p.parseUnary()
				if err != nil {
					return 0, err
				}
				if right == 0 {
					return 0, errors.New("modulo by zero")
				}
				left = math.Mod(left, right)
			} else {
				p.pos = save
				return left, nil
			}
		case p.wordAhead("mod"):
			p.pos += 3
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, errors.New("modulo by zero")
			}
			left = math.Mod(left, right)
		case p.startsImplicitMul():
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			left *= right
		default:
			return left, nil
		}
	}
}

// wordAhead reports whether the next token is exactly the given keyword.
func (p *parser) wordAhead(word string) bool {
	p.skipSpace()
	end := p.pos + len(word)
	if end > len(p.src) {
		return false
	}
	if strings.ToLower(string(p.src[p.pos:end])) != word {
		return false
	}
	return !p.identContinues(end)
}

func (p *parser) startsOperand() bool {
	c := p.peek()
	return c == '(' || unicode.IsDigit(c) || unicode.IsLetter(c) || c == '.' || c == '-' || c == '+'
}

// Implicit multiplication: "2pi", "2(3)", "(1)(2)", "3 sqrt(4)"
func (p *parser) startsImplicitMul() bool {
	c := p.peek()
	return c == '(' || unicode.IsLetter(c)
}

// unary := ('-' | '+') unary | power
func (p *parser) parseUnary() (float64, error) {
	switch p.peek() {
	case '-':
		p.pos++
		v, err := p.parseUnary()
		return -v, err
	case '+':
		p.pos++
		return p.parseUnary()
	}
	return p.parsePower()
}

// power := postfix ('^' unary)?   (right associative)
func (p *parser) parsePower() (float64, error) {
	base, err := p.parsePostfix()
	if err != nil {
		return 0, err
	}
	if p.peek() == '^' {
		p.pos++
		exp, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return math.Pow(base, exp), nil
	}
	if p.pos+1 < len(p.src) && p.src[p.pos] == '*' && p.src[p.pos+1] == '*' {
		p.pos += 2
		exp, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return math.Pow(base, exp), nil
	}
	return base, nil
}

// postfix := primary ('!' | '%')*
func (p *parser) parsePostfix() (float64, error) {
	v, err := p.parsePrimary()
	if err != nil {
		return 0, err
	}
	for {
		switch p.peek() {
		case '!':
			p.pos++
			v, err = factorial(v)
			if err != nil {
				return 0, err
			}
		case '%':
			save := p.pos
			p.pos++
			if p.startsOperand() {
				p.pos = save
				return v, nil
			}
			v /= 100
		default:
			return v, nil
		}
	}
}

func factorial(v float64) (float64, error) {
	if v < 0 || v != math.Trunc(v) {
		return 0, errors.New("factorial needs a non-negative integer")
	}
	if v > 170 {
		return 0, errors.New("factorial too large")
	}
	r := 1.0
	for i := 2.0; i <= v; i++ {
		r *= i
	}
	return r, nil
}

// primary := number | ident | ident '(' args ')' | '(' expr ')'
func (p *parser) parsePrimary() (float64, error) {
	c := p.peek()
	switch {
	case c == '(':
		p.pos++
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.peek() != ')' {
			return 0, errors.New("missing )")
		}
		p.pos++
		return v, nil
	case unicode.IsDigit(c) || c == '.':
		return p.parseNumber()
	case unicode.IsLetter(c):
		return p.parseIdent()
	case c == 0:
		return 0, errors.New("incomplete expression")
	default:
		return 0, fmt.Errorf("unexpected %q", string(c))
	}
}

func (p *parser) parseNumber() (float64, error) {
	start := p.pos
	// Hex, binary, octal literals
	if p.src[p.pos] == '0' && p.pos+1 < len(p.src) {
		var base int
		switch unicode.ToLower(p.src[p.pos+1]) {
		case 'x':
			base = 16
		case 'b':
			base = 2
		case 'o':
			base = 8
		}
		if base != 0 {
			p.pos += 2
			ds := p.pos
			for p.pos < len(p.src) && (unicode.IsDigit(p.src[p.pos]) || unicode.IsLetter(p.src[p.pos]) || p.src[p.pos] == '_') {
				p.pos++
			}
			digits := strings.ReplaceAll(string(p.src[ds:p.pos]), "_", "")
			n, err := strconv.ParseInt(digits, base, 64)
			if err != nil {
				return 0, fmt.Errorf("bad number %q", string(p.src[start:p.pos]))
			}
			return float64(n), nil
		}
	}
	for p.pos < len(p.src) && (unicode.IsDigit(p.src[p.pos]) || p.src[p.pos] == '.' || p.src[p.pos] == '_') {
		p.pos++
	}
	// Scientific notation: 1e5, 2.5E-3 (but not "2e" meaning 2*e)
	if p.pos < len(p.src) && (p.src[p.pos] == 'e' || p.src[p.pos] == 'E') {
		q := p.pos + 1
		if q < len(p.src) && (p.src[q] == '+' || p.src[q] == '-') {
			q++
		}
		if q < len(p.src) && unicode.IsDigit(p.src[q]) {
			for q < len(p.src) && unicode.IsDigit(p.src[q]) {
				q++
			}
			p.pos = q
		}
	}
	text := strings.ReplaceAll(string(p.src[start:p.pos]), "_", "")
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("bad number %q", text)
	}
	// Magnitude suffixes: 2k, 3.5m, 1b
	if p.pos < len(p.src) {
		switch unicode.ToLower(p.src[p.pos]) {
		case 'k':
			if !p.identContinues(p.pos + 1) {
				p.pos++
				v *= 1e3
			}
		case 'm':
			if !p.identContinues(p.pos + 1) {
				p.pos++
				v *= 1e6
			}
		case 'b':
			if !p.identContinues(p.pos + 1) {
				p.pos++
				v *= 1e9
			}
		}
	}
	return v, nil
}

func (p *parser) identContinues(at int) bool {
	return at < len(p.src) && (unicode.IsLetter(p.src[at]) || unicode.IsDigit(p.src[at]))
}

func (p *parser) parseIdent() (float64, error) {
	start := p.pos
	for p.pos < len(p.src) && (unicode.IsLetter(p.src[p.pos]) || unicode.IsDigit(p.src[p.pos])) {
		p.pos++
	}
	name := strings.ToLower(string(p.src[start:p.pos]))

	if p.peek() == '(' {
		p.pos++
		var args []float64
		if p.peek() != ')' {
			for {
				v, err := p.parseExpr()
				if err != nil {
					return 0, err
				}
				args = append(args, v)
				if p.peek() != ',' {
					break
				}
				p.pos++
			}
		}
		if p.peek() != ')' {
			return 0, errors.New("missing )")
		}
		p.pos++
		return callFunc(name, args)
	}

	if v, ok := constants[name]; ok {
		return v, nil
	}
	if name == "mod" {
		return 0, errors.New("mod needs two operands")
	}
	return 0, fmt.Errorf("unknown name %q", name)
}

func callFunc(name string, args []float64) (float64, error) {
	if f, ok := unaryFuncs[name]; ok {
		if len(args) != 1 {
			return 0, fmt.Errorf("%s takes 1 argument", name)
		}
		return f(args[0]), nil
	}
	if f, ok := binaryFuncs[name]; ok {
		if len(args) != 2 {
			return 0, fmt.Errorf("%s takes 2 arguments", name)
		}
		return f(args[0], args[1]), nil
	}
	return 0, fmt.Errorf("unknown function %q", name)
}
