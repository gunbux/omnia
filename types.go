package main

// Variable and Constant Definitions
const (
	MinTerminalWidth     = 40 // Minimum size of the terminal for omnia to function
	MaxBoxWidth          = 80 // Maximum size of the Launcher input box
	CompletionListHeight = 10 // Maximum height of the completion list
)

// Custom Msgs

type focusCompletionMsg struct {
	isCompletionFocused bool
}
