package main

import (
	"os/exec"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	launcherInput  textinput.Model
	completionList list.Model
	windowWidth    int
	windowHeight   int
	userShell      Shell
}

// Bubble Tea Model

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something to launch..."
	ti.Focus()
	ti.CharLimit = 512

	cl := list.New([]list.Item{}, completionDelegate{}, 0, 0)
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
		m.completionList.SetHeight(CompletionListHeight)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		// TODO: Check whether we've selected a completion, and if we have, run that instead? Should we make it run or just fill?
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
