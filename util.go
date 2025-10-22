package main

import (
	"os/exec"
	"strings"
	"syscall"
)

// Generic Helper Functions

// Gets launcher box width from the terminal window width
func getBoxWidth(windowWidth int) int {
	// NOTE: -4 to account for padding and border on both sides
	return min(windowWidth-4, MaxBoxWidth)
}

func runProgram(input string, isTerminal bool) {
	if input == "" {
		return
	}

	// TODO: Support other terminals
	if isTerminal {
		input = "kitty -- " + input
	}

	parts := strings.Fields(input)
	if len(parts) > 0 {
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Start()
	}
}
