// Package completions provides the implementation for any completions for the launcher.
package completions

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	SelectedBorderColor = lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	SelectedTitleColor  = lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}
	SelectedDescColor   = lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	DimmedTitleColor    = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#CCCCCC"}
	DimmedDescColor     = lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#777777"}
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

// CompletionDelegate renders each result on a single line: the title,
// followed by a dimmed description that is truncated to fit.
type CompletionDelegate struct{}

func (cd CompletionDelegate) Height() int                             { return 1 }
func (cd CompletionDelegate) Spacing() int                            { return 0 }
func (cd CompletionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (cd CompletionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(list.DefaultItem)
	if !ok {
		return
	}
	title, desc := i.Title(), i.Description()
	selected := m.Index() == index

	var titleStyle, descStyle lipgloss.Style
	if selected {
		titleStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(SelectedBorderColor).
			Foreground(SelectedTitleColor).
			Bold(true).
			Padding(0, 0, 0, 1)
		descStyle = lipgloss.NewStyle().Foreground(SelectedDescColor)
	} else {
		titleStyle = lipgloss.NewStyle().
			Foreground(DimmedTitleColor).
			Padding(0, 0, 0, 2) //nolint:mnd
		descStyle = lipgloss.NewStyle().Foreground(DimmedDescColor)
	}

	width := m.Width()
	if width <= 0 {
		width = 80
	}
	titleText := ansi.Truncate(title, max(width-4, 8), "…")
	line := titleStyle.Render(titleText)
	if desc != "" {
		remaining := width - lipgloss.Width(line) - 2
		if remaining > 6 {
			line += "  " + descStyle.Render(ansi.Truncate(desc, remaining, "…"))
		}
	}
	fmt.Fprint(w, line)
}
