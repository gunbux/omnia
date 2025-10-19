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

	// TODO: Make this nicer visually
	completionString := string(completionItem)
	if m.Index() == index {
		completionString = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Render(completionString)
	}
	fmt.Fprintln(w, completionString)
}
