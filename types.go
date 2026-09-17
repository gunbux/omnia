package main

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/gunbux/omnia/ai"
	"github.com/gunbux/omnia/currency"
	"github.com/gunbux/omnia/finder"
)

// Mode is what the launcher is currently doing with the input.
type Mode int

const (
	ModeApps Mode = iota
	ModeShell
	ModeFiles
	ModeCalc
	ModeCurrency
	ModeAI
)

func (m Mode) String() string {
	switch m {
	case ModeShell:
		return "Shell"
	case ModeFiles:
		return "Files"
	case ModeCalc:
		return "Calc"
	case ModeCurrency:
		return "Currency"
	case ModeAI:
		return "Ask AI"
	default:
		return "Apps"
	}
}

const (
	MaxBoxWidth        = 80 // Maximum width of the launcher and result boxes
	MaxListHeight      = 12 // Maximum number of visible results
	MinListHeight      = 3
	TerminalCommand    = "kitty"
	FileSearchDebounce = 90 * time.Millisecond
	StatusDuration     = 1500 * time.Millisecond
)

// Custom Msgs

type shellItemsMsg struct{ items []list.Item }

type fileTickMsg struct{ seq int }

type fileResultsMsg struct {
	seq     int
	results []finder.Result
}

type ratesMsg struct {
	rates *currency.Rates
	err   error
}

type aiChunkMsg ai.Chunk

type statusClearMsg struct{}
