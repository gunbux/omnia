package completions

import (
	"fmt"
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
func GetCliCompletionsCmd(input string, shell Shell) tea.Cmd {
	return func() tea.Msg {
		if input == "" {
			return UpdateCompletionMsg{Items: []list.Item{}}
		}

		switch shell {
		case ZSH:
			return UpdateCompletionMsg{Items: getZshCompletions(input)}
		case BASH:
			return UpdateCompletionMsg{Items: getBashCompletions(input)}
		case UNKNOWN:
			fallthrough
		default:
			return UpdateCompletionMsg{Items: []list.Item{}}
		}
	}
}

func getBashCompletions(input string) []list.Item {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("compgen -c '%s'", input))
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

func getZshCompletions(input string) []list.Item {
	script := fmt.Sprintf(`
		setopt NO_NOMATCH
		autoload -U compinit
		compinit -D
		compgen -c '%s' 2>/dev/null || printf '%%s\n' ${(k)commands[(I)%s*]}
	`, input, input)

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
