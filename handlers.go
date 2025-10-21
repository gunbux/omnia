package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// A bunch of handlers for bubble tea Model.

// Function to focus and unfocus the completion list
func handleCompletionFocus(m model, focus bool) (model, tea.Cmd) {
	m.isCompletionFocused = focus
	// delegate := appCompletionDelegate{isCompletionFocused: focus}
	// m.completionList.SetDelegate(delegate)
	return m, nil
}

// Function to handle generic keystroke input. This does not include any special escape keys or interactive keys for the launcher.
func handleGenericKeyInput(keyMsg tea.KeyMsg, m model) (model, tea.Cmd) {
	var inputCmd tea.Cmd
	var completionCmd tea.Cmd
	var focusCompletionCmd tea.Cmd

	// NOTE: When we receive a generic key input, set completion list to unfocus if it is focused.
	if m.isCompletionFocused {
		focusCompletionCmd = func() tea.Msg { return focusCompletionMsg{false} }
	}

	m.launcherInput, inputCmd = m.launcherInput.Update(keyMsg)
	completionCmd = func() tea.Msg { return getAppCompletions() }
	return m, tea.Sequence(focusCompletionCmd, inputCmd, completionCmd)
}

// This represents behaviour when the completion list is focused
func handleMsgCompletionFocused(msg tea.Msg, m model) (model, tea.Cmd) {
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
			if selectedItem := m.completionList.SelectedItem(); selectedItem != nil {
				if entry, ok := selectedItem.(DesktopEntry); ok {
					m.launcherInput.SetValue(entry.exec)
				}
			}
			m.launcherInput.CursorEnd()
			return m, func() tea.Msg { return focusCompletionMsg{false} }
		case tea.KeyEsc:
			return m, func() tea.Msg { return focusCompletionMsg{false} }
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
		return handleGenericKeyInput(msg, m)

	case focusCompletionMsg:
		return handleCompletionFocus(m, msg.isCompletionFocused)
	}
	return m, nil
}

func handleWindowSize(msg tea.WindowSizeMsg, m model) (model, tea.Cmd) {
	m.windowWidth = msg.Width
	m.windowHeight = msg.Height
	boxWidth := getBoxWidth(m.windowWidth)
	m.launcherInput.Width = boxWidth
	m.completionList.SetWidth(boxWidth)
	m.completionList.SetHeight(CompletionListHeight)
	return m, nil
}
