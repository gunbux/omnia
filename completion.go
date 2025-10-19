package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Type representing a completion item.
type completion string

func (c completion) FilterValue() string { return string(c) }
func (c completion) Title() string       { return string(c) }
func (c completion) Description() string { return "" }

// Type defining how the completion list should render
type completionDelegate struct{}

func (cd completionDelegate) Height() int                             { return 1 }
func (cd completionDelegate) Spacing() int                            { return 0 }
func (cd completionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (cd completionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	completionItem, ok := listItem.(completion)
	if !ok {
		return
	}

	completionString := string(completionItem)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		PaddingLeft(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("63")).
		Bold(true).
		PaddingLeft(1)

	prefix := " "
	if m.Index() == index {
		prefix = ">"
		completionString = selectedStyle.Render(fmt.Sprintf("%s %s", prefix, completionString))
	} else {
		completionString = normalStyle.Render(fmt.Sprintf("%s %s", prefix, completionString))
	}

	fmt.Fprint(w, completionString)
}
