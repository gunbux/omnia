package main

import (
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	MinTerminalWidth = 40 // Minimum size of the terminal for omnia to function
	MaxBoxWidth      = 80 // Maximum size of the Launcher input box
)

type model struct {
	launcherInput textinput.Model
	windowWidth   int
	windowHeight  int
}

// Helper Functions
func getBoxWidth(windowWidth int) int {
	// NOTE: -4 to account for padding and border on both sides
	return min(windowWidth-4, MaxBoxWidth)
}

// Bubble Tea Model
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something to launch..."
	ti.Focus()
	ti.CharLimit = 512
	return model{
		launcherInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		boxWidth := getBoxWidth(m.windowWidth)
		m.launcherInput.Width = boxWidth
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	m.launcherInput, cmd = m.launcherInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	boxWidth := getBoxWidth(m.windowWidth)
	launcherBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(boxWidth)

	content := m.launcherInput.View()
	return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, launcherBox.Render(content))
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
