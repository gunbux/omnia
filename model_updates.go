package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func updateCompletionList(msg tea.Msg, m model) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
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
		case tea.KeyEnter:
			m.launcherInput.SetValue(string(m.completionList.SelectedItem().(completion)))
			m.launcherInput.CursorEnd()
			m.isCompletionFocused = false
			return m, nil
		case tea.KeyEsc:
			m.isCompletionFocused = false
			return m, nil
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	return m, nil
}

func updateWindowSize(msg tea.WindowSizeMsg, m model) (model, tea.Cmd) {
	m.windowWidth = msg.Width
	m.windowHeight = msg.Height
	boxWidth := getBoxWidth(m.windowWidth)
	m.launcherInput.Width = boxWidth
	m.completionList.SetWidth(boxWidth)
	m.completionList.SetHeight(CompletionListHeight)
	return m, nil
}
