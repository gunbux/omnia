package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	MinTerminalWidth = 40 // Minimum size of the terminal for omnia to function
	MaxBoxWidth      = 80 // Maximum size of the Launcher input box
)

type Shell int

const (
	ZSH Shell = iota
	BASH
	UNKNOWN
)

type updateCompletionMsg struct {
	completions []list.Item
}

type completion string

func (c completion) FilterValue() string { return string(c) }
func (c completion) Title() string       { return string(c) }
func (c completion) Description() string { return "" }

type model struct {
	launcherInput  textinput.Model
	completionList list.Model
	windowWidth    int
	windowHeight   int
	userShell      Shell
}

// Helper Functions

// Gets launcher box width from the terminal window width
func getBoxWidth(windowWidth int) int {
	// NOTE: -4 to account for padding and border on both sides
	return min(windowWidth-4, MaxBoxWidth)
}

func getUserShell() Shell {
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

// Gets autocompletions based on shell type
func getCompletionsCmd(input string, shell Shell) tea.Cmd {
	return func() tea.Msg {
		if input == "" {
			return updateCompletionMsg{[]list.Item{}}
		}

		switch shell {
		case ZSH:
			return updateCompletionMsg{getZshCompletions(input)}
		case BASH:
			return updateCompletionMsg{getBashCompletions(input)}
		case UNKNOWN:
			fallthrough
		default:
			return updateCompletionMsg{[]list.Item{}}
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
		items[i] = completion(line)
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
		items[i] = completion(line)
	}
	return items
}

// Bubble Tea Model

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something to launch..."
	ti.Focus()
	ti.CharLimit = 512

	cl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	cl.SetShowTitle(false)
	cl.SetShowStatusBar(false)
	cl.SetShowPagination(false)
	cl.SetShowHelp(false)
	cl.SetFilteringEnabled(false)

	return model{
		launcherInput:  ti,
		completionList: cl,
		userShell:      getUserShell(),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		boxWidth := getBoxWidth(m.windowWidth)
		m.launcherInput.Width = boxWidth
		m.completionList.SetWidth(boxWidth)
		// TODO: Make this somewhat dynamic
		m.completionList.SetHeight(m.windowHeight - 10)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			input := strings.TrimSpace(m.launcherInput.Value())
			if input != "" {
				parts := strings.Fields(input)
				if len(parts) > 0 {
					cmd := exec.Command(parts[0], parts[1:]...)
					cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
					cmd.Start()
				}
			}
			return m, tea.Quit
		// TODO: When we select a completion, fill it into the input box without update completionList
		case tea.KeyTab:
			m.completionList.CursorDown()
			return m, nil
		case tea.KeyShiftTab:
			m.completionList.CursorUp()
			return m, nil
		case tea.KeyUp, tea.KeyDown:
			var cmd tea.Cmd
			m.completionList, cmd = m.completionList.Update(msg)
			return m, cmd
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

		// If no early return, the Key will be handled by text input.
		var inputCmd tea.Cmd
		var completionCmd tea.Cmd

		m.launcherInput, inputCmd = m.launcherInput.Update(msg)
		input := strings.TrimSpace(m.launcherInput.Value())
		completionCmd = getCompletionsCmd(input, m.userShell)

		return m, tea.Sequence(inputCmd, completionCmd)
	case updateCompletionMsg:
		m.completionList.SetItems(msg.completions)
		m.completionList.ResetSelected()
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	boxWidth := getBoxWidth(m.windowWidth)

	// Styling
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(boxWidth)

	launcherBox := boxStyle.Render(m.launcherInput.View())
	var completionBox string
	if len(m.completionList.Items()) > 0 {
		completionBox = boxStyle.Render(m.completionList.View())
	}

	content := lipgloss.JoinVertical(lipgloss.Left, launcherBox, completionBox)

	return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, content)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
