package main

// TODO: Put all the different type of completions in a folder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

// TODO: App completions get should be initial, then we just search through this, using filtervalue?

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
