package main

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gunbux/omnia/ai"
	"github.com/gunbux/omnia/calc"
	"github.com/gunbux/omnia/completions"
	"github.com/gunbux/omnia/currency"
	"github.com/gunbux/omnia/finder"
)

// Handlers for the Bubble Tea model.

func handleWindowSize(msg tea.WindowSizeMsg, m model) (model, tea.Cmd) {
	m.windowWidth = msg.Width
	m.windowHeight = msg.Height
	boxWidth := getBoxWidth(m.windowWidth)
	m.resultList.SetWidth(boxWidth - 2)
	m.setPrompt()

	// Launcher box (3 rows) + result box borders (2) + footer (1) + breathing room (2)
	listHeight := m.windowHeight - 8
	listHeight = min(max(listHeight, MinListHeight), MaxListHeight)
	m.resultList.SetHeight(listHeight)

	m.aiView.Width = boxWidth - 2
	m.aiView.Height = listHeight
	if m.aiActive {
		m.updateAIView()
	}
	return m, nil
}

// handleKey handles keys while the launcher input is active.
func handleKey(msg tea.KeyMsg, m model) (model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit

	case tea.KeyEnter:
		return handleEnter(m)

	case tea.KeyUp, tea.KeyShiftTab, tea.KeyCtrlP:
		m.resultList.CursorUp()
		return m, nil

	case tea.KeyDown, tea.KeyTab, tea.KeyCtrlN:
		m.resultList.CursorDown()
		return m, nil

	case tea.KeyRight:
		// At the end of the input, → completes the selected path or command.
		if m.launcherInput.Position() >= len(m.launcherInput.Value()) {
			if completed, ok := completionFor(m); ok {
				m.launcherInput.SetValue(completed)
				m.launcherInput.CursorEnd()
				return refreshResults(m)
			}
		}

	case tea.KeyCtrlY:
		if item, ok := m.resultList.SelectedItem().(copyable); ok {
			return copyAndNotify(m, item.Text(), false)
		}
		return m, nil

	case tea.KeyCtrlT:
		if m.mode == ModeFiles {
			if item, ok := m.resultList.SelectedItem().(fileItem); ok {
				openTerminalAt(dirOf(item))
				return m, tea.Quit
			}
		}
		return m, nil
	}

	before := m.launcherInput.Value()
	var inputCmd tea.Cmd
	m.launcherInput, inputCmd = m.launcherInput.Update(msg)
	if m.launcherInput.Value() == before {
		return m, inputCmd
	}
	m, refreshCmd := refreshResults(m)
	return m, tea.Batch(inputCmd, refreshCmd)
}

// completionFor returns the text to put in the input when → is pressed.
func completionFor(m model) (string, bool) {
	switch item := m.resultList.SelectedItem().(type) {
	case fileItem:
		path := finder.Abbreviate(item.Path)
		if item.IsDir {
			path += "/"
		}
		if strings.HasPrefix(m.launcherInput.Value(), "f ") {
			return "f " + path, true
		}
		return path, true
	case completions.ShellCompletionEntry:
		return "> " + string(item) + " ", true
	}
	return "", false
}

func dirOf(item fileItem) string {
	if item.IsDir {
		return item.Path
	}
	return strings.TrimSuffix(item.Path, "/"+item.Name())
}

// handleEnter performs the primary action for the current mode.
func handleEnter(m model) (model, tea.Cmd) {
	selected := m.resultList.SelectedItem()

	switch m.mode {
	case ModeApps:
		if entry, ok := selected.(completions.DesktopEntry); ok {
			runProgram(entry.Exec, entry.Terminal)
			return m, tea.Quit
		}
		// Nothing matched: treat the input as a command to launch.
		runProgram(m.query, false)
		return m, tea.Quit

	case ModeShell:
		command := m.query
		// A bare word takes the highlighted completion; anything with
		// arguments runs exactly as typed.
		if entry, ok := selected.(completions.ShellCompletionEntry); ok && !strings.Contains(command, " ") {
			command = string(entry)
		}
		if command == "" {
			return m, nil
		}
		runProgram(command, true)
		return m, tea.Quit

	case ModeFiles:
		if item, ok := selected.(fileItem); ok {
			openPath(item.Path)
			return m, tea.Quit
		}
		return m, nil

	case ModeCalc, ModeCurrency:
		if item, ok := selected.(copyable); ok {
			return copyAndNotify(m, item.Text(), true)
		}
		return m, nil

	case ModeAI:
		if m.query == "" {
			return m, nil
		}
		return startAI(m, m.query)
	}
	return m, nil
}

func copyAndNotify(m model, text string, quitAfter bool) (model, tea.Cmd) {
	if err := copyToClipboard(text); err != nil {
		m.status = "Copy failed: " + err.Error()
		return m, clearStatusLater()
	}
	if quitAfter {
		return m, tea.Quit
	}
	m.status = "Copied " + truncate(text, 50)
	return m, clearStatusLater()
}

func clearStatusLater() tea.Cmd {
	return tea.Tick(StatusDuration, func(time.Time) tea.Msg { return statusClearMsg{} })
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// refreshResults recomputes the mode and result list from the current input.
func refreshResults(m model) (model, tea.Cmd) {
	prevMode := m.mode
	mode, query := detectMode(m.launcherInput.Value())
	m.mode, m.query = mode, query
	m.setPrompt()

	var cmd tea.Cmd
	switch mode {
	case ModeApps:
		setItems(&m, m.apps)
		if query != "" {
			m.resultList.SetFilterText(query)
		}

	case ModeShell:
		if !m.shellLoaded {
			if !m.shellAsked {
				m.shellAsked = true
				cmd = func() tea.Msg {
					return shellItemsMsg{items: completions.GetCliCompletionsCmd().(completions.UpdateCompletionItemsMsg).Items}
				}
			}
			setItems(&m, []list.Item{infoItem{title: "Loading commands…"}})
			break
		}
		setItems(&m, m.shellCmds)
		if word, _, _ := strings.Cut(query, " "); word != "" {
			m.resultList.SetFilterText(word)
		}

	case ModeFiles:
		m.searchSeq++
		if query == "" {
			setItems(&m, []list.Item{infoItem{title: "Type a file name", desc: "or a path like ~/Documents/"}})
			break
		}
		if prevMode != ModeFiles {
			setItems(&m, []list.Item{infoItem{title: "Searching…"}})
		}
		seq := m.searchSeq
		cmd = tea.Tick(FileSearchDebounce, func(time.Time) tea.Msg { return fileTickMsg{seq: seq} })

	case ModeCalc:
		setItems(&m, calcItems(query))

	case ModeCurrency:
		items, fetch := currencyItems(&m, query)
		setItems(&m, items)
		cmd = fetch

	case ModeAI:
		if query == "" {
			setItems(&m, []list.Item{infoItem{title: "Ask anything", desc: "e.g. ? how do I untar a .tar.xz"}})
		} else {
			setItems(&m, []list.Item{aiItem{prompt: query}})
		}
	}
	return m, cmd
}

// setItems replaces the list contents with no filter applied.
func setItems(m *model, items []list.Item) {
	m.resultList.ResetFilter()
	m.resultList.SetItems(items)
	m.resultList.ResetSelected()
}

func calcItems(query string) []list.Item {
	if query == "" {
		return []list.Item{infoItem{title: "Type an expression", desc: "e.g. 2^10, sqrt(2)*pi, 15% of 80 → 80*15%"}}
	}
	v, err := calc.Eval(query)
	if err != nil {
		return []list.Item{infoItem{title: "…", desc: err.Error()}}
	}
	return []list.Item{calcItem{value: v}}
}

func currencyItems(m *model, query string) ([]list.Item, tea.Cmd) {
	q, ok := currency.Parse(query)
	if !ok {
		return []list.Item{infoItem{title: "Type a conversion", desc: "e.g. 100 usd to eur"}}, nil
	}

	if m.rates == nil {
		if m.ratesErr != nil {
			return []list.Item{infoItem{title: "Rates unavailable", desc: m.ratesErr.Error()}}, nil
		}
		var cmd tea.Cmd
		if !m.ratesLoading {
			m.ratesLoading = true
			cmd = func() tea.Msg {
				r, err := currency.Load()
				return ratesMsg{rates: r, err: err}
			}
		}
		return []list.Item{infoItem{title: "Fetching exchange rates…"}}, cmd
	}

	targets := []string{q.To}
	if q.To == "" {
		targets = targets[:0]
		for _, t := range currency.DefaultTargets {
			if t != q.From {
				targets = append(targets, t)
			}
		}
	}

	var items []list.Item
	for _, to := range targets {
		result, err := m.rates.Convert(q.Amount, q.From, to)
		if err != nil {
			items = append(items, infoItem{title: "Unknown currency " + to})
			continue
		}
		rate, _ := m.rates.Convert(1, q.From, to)
		items = append(items, currencyItem{amount: q.Amount, from: q.From, to: to, result: result, rate: rate})
	}
	if m.rates.Stale() && m.rates.Updated != "" {
		items = append(items, infoItem{title: "Rates from " + m.rates.Updated, desc: "offline, could not refresh"})
	}
	return items, nil
}

func searchFilesCmd(query string, seq int) tea.Cmd {
	return func() tea.Msg {
		return fileResultsMsg{seq: seq, results: finder.Search(query, finder.DefaultLimit)}
	}
}

// AI answer view

func startAI(m model, prompt string) (model, tea.Cmd) {
	ctx, cancel := context.WithCancel(context.Background())
	m.aiActive = true
	m.aiPrompt = prompt
	m.aiAnswer.Reset()
	m.aiDone = false
	m.aiErr = nil
	m.aiCancel = cancel
	m.aiChan = ai.Ask(ctx, prompt)
	m.launcherInput.Blur()
	m.updateAIView()
	return m, tea.Batch(waitForChunk(m.aiChan), m.aiSpin.Tick)
}

func waitForChunk(ch <-chan ai.Chunk) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return aiChunkMsg{Done: true}
		}
		return aiChunkMsg(chunk)
	}
}

func stopAI(m model) model {
	if m.aiCancel != nil {
		m.aiCancel()
	}
	m.aiActive = false
	m.aiCancel = nil
	m.aiChan = nil
	m.launcherInput.Focus()
	return m
}

func handleAIKey(msg tea.KeyMsg, m model) (model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return stopAI(m), tea.Quit
	case tea.KeyEsc:
		m = stopAI(m)
		return m, textinput.Blink
	case tea.KeyEnter, tea.KeyCtrlY:
		answer := strings.TrimSpace(m.aiAnswer.String())
		if answer == "" {
			return m, nil
		}
		return copyAndNotify(m, answer, msg.Type == tea.KeyEnter)
	}
	var cmd tea.Cmd
	m.aiView, cmd = m.aiView.Update(msg)
	return m, cmd
}
