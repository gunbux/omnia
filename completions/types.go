// Package completions provides the implementation for any completions for the launcher.
package completions

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NOTE: I've split up the updates just because the current mental model
// of how we should do completions is pregenerating a list of completions the
// filtering off that. This may not work if we get into contextual completions.

type UpdateCompletionItemsMsg struct {
	Items []list.Item
}

type UpdateCompletionFilterMsg struct {
	Input string
}

type CompletionDelegate struct {
	IsCompletionFocused bool
}

func (cd CompletionDelegate) Height() int                             { return 1 }
func (cd CompletionDelegate) Spacing() int                            { return 0 }
func (cd CompletionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (cd CompletionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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

	if m.Index() == index && cd.IsCompletionFocused {
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
