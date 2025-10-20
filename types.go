package main

import (
	"github.com/charmbracelet/bubbles/list"
)

// Variable and Constant Definitions
const (
	MinTerminalWidth     = 40 // Minimum size of the terminal for omnia to function
	MaxBoxWidth          = 80 // Maximum size of the Launcher input box
	CompletionListHeight = 10 // Maximum height of the completion list
)

type Shell int

const (
	ZSH Shell = iota
	BASH
	UNKNOWN
)

// Custom Msgs

type updateCompletionMsg struct {
	completions []list.Item
}

type focusCompletionMsg struct {
	isCompletionFocused bool
}
