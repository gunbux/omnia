package main

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gunbux/omnia/ai"
	"github.com/gunbux/omnia/completions"
	"github.com/gunbux/omnia/currency"
)

const (
	BorderColorFocused   = "63"
	BorderColorUnfocused = "8"   // Brighter grey (standard terminal bright black)
	TextColorUnfocused   = "245" // Lighter grey for better readability
	TextColorHint        = "241"
	TextColorStatus      = "78"
	TextColorError       = "203"
)

var modeChipColors = map[Mode]string{
	ModeApps:     "63",
	ModeShell:    "208",
	ModeFiles:    "39",
	ModeCalc:     "78",
	ModeCurrency: "220",
	ModeAI:       "170",
}

type model struct {
	launcherInput textinput.Model
	resultList    list.Model
	windowWidth   int
	windowHeight  int

	mode  Mode
	query string // input with the mode prefix stripped

	apps        []list.Item
	shellCmds   []list.Item
	shellLoaded bool
	shellAsked  bool

	rates        *currency.Rates
	ratesLoading bool
	ratesErr     error

	searchSeq int // bumps on every keystroke in file mode to drop stale results

	status string // transient footer message, e.g. "Copied"

	// AI answer view
	aiActive bool
	aiPrompt string
	aiAnswer strings.Builder
	aiDone   bool
	aiErr    error
	aiCancel context.CancelFunc
	aiChan   <-chan ai.Chunk
	aiView   viewport.Model
	aiSpin   spinner.Model
}

// Bubble Tea Model

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Search apps, or try  = 2+2   ? ask   > cmd   / files"
	ti.Focus()
	ti.CharLimit = 512

	rl := list.New([]list.Item{}, completions.CompletionDelegate{}, 0, 0)
	rl.SetShowTitle(false)
	rl.SetShowStatusBar(false)
	rl.SetShowPagination(false)
	rl.SetShowHelp(false)
	rl.SetFilteringEnabled(true)
	rl.SetShowFilter(false)
	rl.DisableQuitKeybindings()

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(modeChipColors[ModeAI]))

	m := model{
		launcherInput: ti,
		resultList:    rl,
		aiSpin:        sp,
		aiView:        viewport.New(0, 0),
	}
	m.setPrompt()
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		func() tea.Msg { return completions.GetDesktopCompletions() },
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return handleWindowSize(msg, m)

	case tea.KeyMsg:
		if m.aiActive {
			return handleAIKey(msg, m)
		}
		return handleKey(msg, m)

	// Data arriving for the result list
	case completions.UpdateCompletionItemsMsg:
		m.apps = msg.Items
		if m.mode == ModeApps {
			return refreshResults(m)
		}
		return m, nil

	case shellItemsMsg:
		m.shellCmds = msg.items
		m.shellLoaded = true
		if m.mode == ModeShell {
			return refreshResults(m)
		}
		return m, nil

	case fileTickMsg:
		if msg.seq != m.searchSeq || m.mode != ModeFiles {
			return m, nil
		}
		return m, searchFilesCmd(m.query, msg.seq)

	case fileResultsMsg:
		if msg.seq != m.searchSeq || m.mode != ModeFiles {
			return m, nil
		}
		items := make([]list.Item, 0, len(msg.results))
		for _, r := range msg.results {
			items = append(items, fileItem{r})
		}
		if len(items) == 0 {
			items = []list.Item{infoItem{title: "No matches", desc: "try a different name or a path like ~/Documents/"}}
		}
		setItems(&m, items)
		return m, nil

	case ratesMsg:
		m.ratesLoading = false
		m.rates, m.ratesErr = msg.rates, msg.err
		if m.mode == ModeCurrency {
			return refreshResults(m)
		}
		return m, nil

	// AI streaming
	case aiChunkMsg:
		if !m.aiActive {
			return m, nil
		}
		m.aiAnswer.WriteString(msg.Text)
		m.aiErr = msg.Err
		if msg.Done {
			m.aiDone = true
			m.updateAIView()
			return m, nil
		}
		m.updateAIView()
		return m, waitForChunk(m.aiChan)

	case spinner.TickMsg:
		if m.aiActive && !m.aiDone {
			var cmd tea.Cmd
			m.aiSpin, cmd = m.aiSpin.Update(msg)
			m.updateAIView()
			return m, cmd
		}
		return m, nil

	case statusClearMsg:
		m.status = ""
		return m, nil
	}

	return m, nil
}

// setPrompt renders the mode chip into the text input's prompt.
func (m *model) setPrompt() {
	chip := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color(modeChipColors[m.mode])).
		Padding(0, 1).
		Render(m.mode.String())
	m.launcherInput.Prompt = chip + " "
	// Text area is the box interior minus padding and the chip.
	m.launcherInput.Width = max(getBoxWidth(m.windowWidth)-4-lipgloss.Width(m.launcherInput.Prompt), 10)
}

func (m model) View() string {
	boxWidth := getBoxWidth(m.windowWidth)

	borderColor := BorderColorFocused
	if m.aiActive {
		borderColor = BorderColorUnfocused
	}
	launcherBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Width(boxWidth)

	resultBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColorUnfocused)).
		Padding(0, 1).
		Width(boxWidth)

	var body string
	if m.aiActive {
		resultBoxStyle = resultBoxStyle.BorderForeground(lipgloss.Color(modeChipColors[ModeAI]))
		body = m.aiView.View()
	} else {
		body = m.resultList.View()
	}

	launcherBox := launcherBoxStyle.Render(m.launcherInput.View())
	resultBox := resultBoxStyle.Render(body)

	var footerText string
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(TextColorHint)).Width(boxWidth+2).Padding(0, 1)
	switch {
	case m.status != "":
		footerText = m.status
		footerStyle = footerStyle.Foreground(lipgloss.Color(TextColorStatus))
	case m.aiActive && m.aiErr != nil:
		footerText = "error: " + m.aiErr.Error()
		footerStyle = footerStyle.Foreground(lipgloss.Color(TextColorError))
	case m.aiActive:
		footerText = "enter copy answer & close  ·  ↑↓ scroll  ·  esc back"
	default:
		footerText = modeHints(m.mode)
	}
	footer := footerStyle.Render(footerText)

	content := lipgloss.JoinVertical(lipgloss.Left, launcherBox, resultBox, footer)
	return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, content)
}

// updateAIView re-renders the streamed answer into the viewport, keeping the
// bottom visible while text is still arriving.
func (m *model) updateAIView() {
	width := max(m.aiView.Width, 10)
	var content string
	switch {
	case m.aiAnswer.Len() == 0 && !m.aiDone:
		content = m.aiSpin.View() + " Thinking…"
	case m.aiAnswer.Len() == 0 && m.aiErr != nil:
		content = lipgloss.NewStyle().Foreground(lipgloss.Color(TextColorError)).Render("No answer: " + m.aiErr.Error())
	default:
		content = lipgloss.NewStyle().Width(width).Render(strings.TrimSpace(m.aiAnswer.String()))
	}
	atBottom := m.aiView.AtBottom()
	m.aiView.SetContent(content)
	if atBottom || !m.aiDone {
		m.aiView.GotoBottom()
	}
}
