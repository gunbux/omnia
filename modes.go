package main

import (
	"strings"

	"github.com/gunbux/omnia/calc"
	"github.com/gunbux/omnia/currency"
)

// detectMode decides what the input means and returns the mode plus the
// query with any prefix stripped.
//
// Explicit prefixes always win:
//
//	?  ask AI      =  calculator     >  shell command
//	/  ~  file path    f  file search
//
// Without a prefix, currency conversions ("100 usd to eur") and arithmetic
// ("2*(3+4)") are recognised automatically; anything else searches apps.
func detectMode(input string) (Mode, string) {
	s := strings.TrimSpace(input)
	if s == "" {
		return ModeApps, ""
	}

	switch {
	case strings.HasPrefix(s, "?"):
		return ModeAI, strings.TrimSpace(s[1:])
	case strings.HasPrefix(s, "="):
		return ModeCalc, strings.TrimSpace(s[1:])
	case strings.HasPrefix(s, ">"):
		return ModeShell, strings.TrimSpace(s[1:])
	case strings.HasPrefix(s, "/") || strings.HasPrefix(s, "~"):
		return ModeFiles, s
	case s == "f":
		return ModeFiles, ""
	case strings.HasPrefix(s, "f "):
		return ModeFiles, strings.TrimSpace(s[2:])
	case strings.HasPrefix(strings.ToLower(s), "ai "):
		return ModeAI, strings.TrimSpace(s[3:])
	}

	if _, ok := currency.Parse(s); ok {
		return ModeCurrency, s
	}
	if calc.LooksLikeExpression(s) {
		return ModeCalc, s
	}
	return ModeApps, s
}

// modeHints returns the footer key hints for a mode.
func modeHints(m Mode) string {
	switch m {
	case ModeShell:
		return "enter run in terminal  ·  → complete  ·  esc quit"
	case ModeFiles:
		return "enter open  ·  → complete path  ·  ^t terminal here  ·  ^y copy path"
	case ModeCalc:
		return "enter copy result  ·  esc quit"
	case ModeCurrency:
		return "enter copy amount  ·  ↑↓ choose currency  ·  esc quit"
	case ModeAI:
		return "enter ask  ·  esc quit"
	default:
		return "?  ask   =  calc   >  shell   /  files   100 usd to eur"
	}
}
