package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gunbux/omnia/completions"
)

const (
	BorderColorFocused   = "63"
	BorderColorUnfocused = "8"   // Brighter grey (standard terminal bright black)
	TextColorUnfocused   = "245" // Lighter grey for better readability
)

type model struct {
	launcherInput       textinput.Model
	completionList      list.Model
	isCompletionFocused bool
	windowWidth         int
	windowHeight        int
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
	cl.SetFilteringEnabled(true)
	cl.SetShowFilter(false)

	return model{
		launcherInput:       ti,
		completionList:      cl,
		isCompletionFocused: false,
	}
}

func (m model) Init() tea.Cmd {
	var completionCmd func() tea.Msg
	switch DefaultCompletionMode {
	case CliCompletionMode:
		completionCmd = func() tea.Msg { return completions.GetCliCompletionsCmd() }
	case DesktopCompletionMode:
		completionCmd = func() tea.Msg { return completions.GetDesktopCompletions() }
	default:
		completionCmd = func() tea.Msg { return completions.GetDesktopCompletions() }
	}

	return tea.Batch(
		textinput.Blink,
		completionCmd,
	)
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
			// TODO: This behaviour is kinda hacky, is there a better way to write this?
			handleQuickRun(m, input)
			return m, tea.Quit
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyUp, tea.KeyDown:
			return m, func() tea.Msg { return focusCompletionMsg{true} }
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

		// If no early return, the Key will be handled by text input.
		return handleGenericKeyInput(msg, m)

	// Custom Msgs
	case completions.UpdateCompletionItemsMsg:
		m.completionList.SetItems(msg.Items)
		m.completionList.ResetSelected()
		return m, nil

	case completions.UpdateCompletionFilterMsg:
		m.completionList.SetFilterText(msg.Input)
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
	var launcherBoxStyle lipgloss.Style
	if !m.isCompletionFocused {
		launcherBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderColorFocused)).
			Bold(true).
			Padding(1).
			Width(boxWidth)
	} else {
		launcherBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderColorUnfocused)).
			Foreground(lipgloss.Color(TextColorUnfocused)).
			Padding(1).
			Width(boxWidth)
	}

	var completionBoxStyle lipgloss.Style
	if m.isCompletionFocused {
		completionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderColorFocused)).
			Bold(true).
			Padding(1).
			Width(boxWidth)
	} else {
		completionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderColorUnfocused)).
			Foreground(lipgloss.Color(TextColorUnfocused)).
			Padding(1).
			Width(boxWidth)
	}

	launcherBox := launcherBoxStyle.Render(m.launcherInput.View())
	completionBox := completionBoxStyle.Render(m.completionList.View())
	content := lipgloss.JoinVertical(lipgloss.Left, launcherBox, completionBox)

	return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, content)
}
