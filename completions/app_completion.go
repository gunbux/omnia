package completions

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
	ID   string
	Desc string
	Exec string
	Icon string // NOTE: this should be a union type of either a name or filepath
}

func (d DesktopEntry) FilterValue() string { return d.ID }
func (d DesktopEntry) Title() string       { return d.ID }
func (d DesktopEntry) Description() string { return d.Desc }

// TODO: App completions get should be initial, then we just search through this, using filtervalue?

// GetAppCompletions based on applications
func GetAppCompletions() tea.Msg {
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

			exec = strings.ReplaceAll(exec, "%U", "")
			exec = strings.ReplaceAll(exec, "%f", "")
			exec = strings.ReplaceAll(exec, "%F", "")
			exec = strings.ReplaceAll(exec, "%u", "")
			exec = strings.ReplaceAll(exec, "%i", "")
			exec = strings.ReplaceAll(exec, "%c", "")
			exec = strings.ReplaceAll(exec, "%k", "")
			exec = strings.ReplaceAll(exec, "%d", "")
			exec = strings.ReplaceAll(exec, "%D", "")
			exec = strings.ReplaceAll(exec, "%N", "")
			exec = strings.ReplaceAll(exec, "%n", "")

			desktopEntry := DesktopEntry{
				ID:   name,
				Desc: description,
				Exec: exec,
				Icon: icon,
			}

			items = append(items, desktopEntry)
		}
	}

	return UpdateCompletionMsg{Items: items}
}
