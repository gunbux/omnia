package main

import (
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	launcherInput textinput.Model
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something to launch..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
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
	return m.launcherInput.View()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
