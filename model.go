package main

import (
	"os/exec"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gunbux/omnia/completions"
)

type model struct {
	launcherInput       textinput.Model
	completionList      list.Model
	isCompletionFocused bool
	windowWidth         int
	windowHeight        int
	userShell           completions.Shell
}

// Bubble Tea Model

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something to launch..."
	ti.Focus()
	ti.CharLimit = 512

	cl := list.New([]list.Item{}, completions.CompletionDelegate{IsCompletionFocused: false}, 0, 0)
	cl.SetShowTitle(false)
	cl.SetShowStatusBar(false)
	cl.SetShowPagination(false)
	cl.SetShowHelp(false)
	cl.SetFilteringEnabled(false)

	return model{
		launcherInput:       ti,
		completionList:      cl,
		isCompletionFocused: false,
		userShell:           completions.GetUserShell(),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Completion Focused Behaviour
	if m.isCompletionFocused {
		return handleMsgCompletionFocused(msg, m)
	}

	// Default Behaviour
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return handleWindowSize(msg, m)
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
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyUp, tea.KeyDown:
			return m, func() tea.Msg { return focusCompletionMsg{true} }
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

		// If no early return, the Key will be handled by text input.
		return handleGenericKeyInput(msg, m)

	// Custom Msgs
	case completions.UpdateCompletionMsg:
		m.completionList.SetItems(msg.Items)
		m.completionList.ResetSelected()
		return m, nil

	case focusCompletionMsg:
		return handleCompletionFocus(m, msg.isCompletionFocused)
	}

	return m, nil
}

func (m model) View() string {
	boxWidth := getBoxWidth(m.windowWidth)

	// Styling
	launcherBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(boxWidth)

	var completionBoxStyle lipgloss.Style
	if m.isCompletionFocused {
		// Focused: keep current color
		completionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1).
			Width(boxWidth)
	} else {
		// Unfocused: grey out
		completionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.Color("240")).
			Padding(1).
			Width(boxWidth)
	}

	launcherBox := launcherBoxStyle.Render(m.launcherInput.View())
	completionBox := completionBoxStyle.Render(m.completionList.View())
	content := lipgloss.JoinVertical(lipgloss.Left, launcherBox, completionBox)

	return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, content)
}
