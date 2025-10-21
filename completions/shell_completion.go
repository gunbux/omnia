package completions

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ShellCompletionEntry string

func (c ShellCompletionEntry) FilterValue() string { return string(c) }
func (c ShellCompletionEntry) Title() string       { return string(c) }
func (c ShellCompletionEntry) Description() string { return "" }

type Shell int

const (
	ZSH Shell = iota
	BASH
	UNKNOWN
)

func GetUserShell() Shell {
	shell := os.Getenv("SHELL")

	// Fallback to $0
	if shell == "" {
		shell = os.Getenv("0")
	}

	shellName := strings.ToLower(filepath.Base(shell))

	switch shellName {
	case "zsh":
		return ZSH
	case "bash":
		return BASH
	default:
		return UNKNOWN
	}
}

// GetCliCompletionsCmd gets completions based on shell type
func GetCliCompletionsCmd() tea.Msg {
	shell := GetUserShell()
	switch shell {
	case ZSH:
		return UpdateCompletionItemsMsg{Items: getZshCompletions()}
	case BASH:
		return UpdateCompletionItemsMsg{Items: getBashCompletions()}
	case UNKNOWN:
		fallthrough
	default:
		return UpdateCompletionItemsMsg{Items: []list.Item{}}
	}
}

func getBashCompletions() []list.Item {
	cmd := exec.Command("bash", "-c", "compgen -c")
	output, err := cmd.Output()
	if err != nil {
		return []list.Item{}
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []list.Item{}
	}

	items := make([]list.Item, len(lines))
	for i, line := range lines {
		items[i] = ShellCompletionEntry(line)
	}
	return items
}

func getZshCompletions() []list.Item {
	script := `
		setopt NO_NOMATCH
		autoload -U compinit
		compinit -D
		print -l ${(k)commands} ${(k)builtins} ${(k)functions} ${(k)aliases} ${(k)reswords}
	`

	cmd := exec.Command("zsh", "-c", script)
	output, err := cmd.Output()
	if err != nil {
		return []list.Item{}
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []list.Item{}
	}

	items := make([]list.Item, len(lines))
	for i, line := range lines {
		items[i] = ShellCompletionEntry(line)
	}
	return items
}
