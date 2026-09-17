package main

import (
	"errors"
	"os/exec"
	"strings"
	"syscall"

	"github.com/atotto/clipboard"
)

// Generic Helper Functions

// Gets launcher box width from the terminal window width
func getBoxWidth(windowWidth int) int {
	// NOTE: -4 to account for padding and border on both sides
	return min(windowWidth-4, MaxBoxWidth)
}

// runProgram launches a program detached from the launcher so it survives
// the launcher exiting. When isTerminal is set the program runs inside a
// new terminal window.
func runProgram(input string, isTerminal bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	var cmd *exec.Cmd
	if isTerminal {
		// Run through the shell so pipes, quotes and aliases work.
		cmd = exec.Command(TerminalCommand, "--", "sh", "-c", input)
	} else {
		parts := strings.Fields(input)
		cmd = exec.Command(parts[0], parts[1:]...)
	}
	startDetached(cmd)
}

// openPath opens a file or directory with the desktop's default handler.
func openPath(path string) {
	startDetached(exec.Command("xdg-open", path))
}

// openTerminalAt opens a new terminal window in the given directory.
func openTerminalAt(dir string) {
	startDetached(exec.Command(TerminalCommand, "--directory", dir))
}

func startDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	_ = cmd.Start()
}

// copyToClipboard tries Wayland and X11 clipboard tools before falling back
// to the OSC52-free library implementation.
func copyToClipboard(text string) error {
	for _, tool := range [][]string{
		{"wl-copy"},
		{"xclip", "-selection", "clipboard"},
		{"xsel", "--clipboard", "--input"},
	} {
		if _, err := exec.LookPath(tool[0]); err != nil {
			continue
		}
		cmd := exec.Command(tool[0], tool[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	if err := clipboard.WriteAll(text); err != nil {
		return errors.New("no clipboard tool found (install wl-clipboard or xclip)")
	}
	return nil
}
