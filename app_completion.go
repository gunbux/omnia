package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/ini.v1"
)

type DesktopEntry struct {
	id          string
	description string
	exec        string
	icon        string // NOTE: this should be a union type of either a name or filepath
}

func (d DesktopEntry) FilterValue() string { return d.id }
func (d DesktopEntry) Title() string       { return d.id }
func (d DesktopEntry) Description() string { return d.description }

// Type defining how the app completion list should render
type appCompletionDelegate struct {
	isCompletionFocused bool
}

func (acd appCompletionDelegate) Height() int                             { return 1 }
func (acd appCompletionDelegate) Spacing() int                            { return 0 }
func (acd appCompletionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (acd appCompletionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	desktopEntry, ok := listItem.(DesktopEntry)
	if !ok {
		return
	}

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		PaddingLeft(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("63")).
		Bold(true).
		PaddingLeft(1)

	descriptionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	prefix := " "
	var renderedContent string

	if desktopEntry.description != "" {
		content := fmt.Sprintf("%s - %s", desktopEntry.id, desktopEntry.description)
		if m.Index() == index && acd.isCompletionFocused {
			prefix = ">"
			renderedContent = selectedStyle.Render(fmt.Sprintf("%s %s", prefix, content))
		} else {
			nameDesc := fmt.Sprintf("%s - ", desktopEntry.id)
			styledName := normalStyle.Render(fmt.Sprintf("%s %s", prefix, nameDesc))
			styledDesc := descriptionStyle.Render(desktopEntry.description)
			renderedContent = styledName + styledDesc
		}
	} else {
		if m.Index() == index && acd.isCompletionFocused {
			prefix = ">"
			renderedContent = selectedStyle.Render(fmt.Sprintf("%s %s", prefix, desktopEntry.id))
		} else {
			renderedContent = normalStyle.Render(fmt.Sprintf("%s %s", prefix, desktopEntry.id))
		}
	}

	fmt.Fprint(w, renderedContent)
}

// Get autocompletions based on applications
func getAppCompletions() tea.Msg {
	var items []list.Item

	for _, dir := range xdg.ApplicationDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}

			desktopFile := filepath.Join(dir, entry.Name())
			cfg, err := ini.Load(desktopFile)
			if err != nil {
				continue
			}

			desktopSection := cfg.Section("Desktop Entry")
			if desktopSection == nil {
				continue
			}

			name := desktopSection.Key("Name").String()
			description := desktopSection.Key("Comment").String()
			exec := desktopSection.Key("Exec").String()
			icon := desktopSection.Key("Icon").String()

			if name == "" || exec == "" {
				continue
			}

			replaceInString(&exec, "%U", "")
			replaceInString(&exec, "%f", "")
			replaceInString(&exec, "%F", "")
			replaceInString(&exec, "%u", "")
			replaceInString(&exec, "%i", "")
			replaceInString(&exec, "%c", "")
			replaceInString(&exec, "%k", "")
			replaceInString(&exec, "%d", "")
			replaceInString(&exec, "%D", "")
			replaceInString(&exec, "%N", "")
			replaceInString(&exec, "%n", "")

			desktopEntry := DesktopEntry{
				id:          name,
				description: description,
				exec:        exec,
				icon:        icon,
			}

			items = append(items, desktopEntry)
		}
	}

	return updateCompletionMsg{items}
}
