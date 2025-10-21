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
type completionDelegate struct {
	isCompletionFocused bool
}

// TODO: Make this a generic delegate

func (cd completionDelegate) Height() int                             { return 1 }
func (cd completionDelegate) Spacing() int                            { return 0 }
func (cd completionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (cd completionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	// Styles
	// NormalStyle := lipgloss.NewStyle().
	// 	Foreground(lipgloss.Color("250")).
	// 	PaddingLeft(1)
	SelectedTitle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 0, 0, 1)

	SelectedDesc := SelectedTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})

	DimmedTitle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 0, 2) //nolint:mnd

	DimmedDesc := DimmedTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})

	var title, desc string

	if i, ok := listItem.(list.DefaultItem); ok {
		title = i.Title()
		desc = i.Description()
	} else {
		return
	}

	if m.Index() == index && cd.isCompletionFocused {
		title = SelectedTitle.Render(title)
		if desc != "" {
			desc = SelectedDesc.Render(desc)
		}
	} else {
		title = DimmedTitle.Render(title)
		if desc != "" {
			desc = DimmedDesc.Render(desc)
		}
	}

	if desc != "" {
		fmt.Fprintf(w, "%s\n%s", title, desc)
		return
	}
	fmt.Fprintf(w, "%s", title)
}
