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
	ID          string
	Desc        string
	Exec        string
	Icon        string // NOTE: this should be a union type of either a name or filepath
	Terminal    bool
	DesktopFile string
}

func (d DesktopEntry) FilterValue() string { return d.ID }
func (d DesktopEntry) Title() string       { return d.ID }
func (d DesktopEntry) Description() string { return d.Desc }

// GetDesktopCompletions based on applications
func GetDesktopCompletions() tea.Msg {
	var items []list.Item
	seen := make(map[string]bool)

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
			// TODO: Proper error handling, along all the catchall errors
			terminal, _ := desktopSection.Key("Terminal").Bool()

			if name == "" || exec == "" {
				continue
			}

			if seen[name] {
				continue
			}
			seen[name] = true

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
				ID:          name,
				Desc:        description,
				Exec:        exec,
				Icon:        icon,
				Terminal:    terminal,
				DesktopFile: desktopFile,
			}

			items = append(items, desktopEntry)
		}
	}

	return UpdateCompletionItemsMsg{Items: items}
}
