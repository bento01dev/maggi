package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	green  = lipgloss.Color("#04b575")
	yellow = lipgloss.Color("#ffd866")
	red    = lipgloss.Color("#ff6188")
	blue   = lipgloss.Color("#2ea0f9")
	muted  = lipgloss.Color("241")
)

type pageType int

const (
	splash pageType = iota
	issue
)

type Page interface {
	tea.Model
	UpdateSize(width, height int)
}

type MaggiModel struct {
	currentPage pageType
	pages       map[pageType]Page
	err         error
}

func NewMaggiModel(debugFlag bool) *MaggiModel {
	return &MaggiModel{
		pages: map[pageType]Page{
			issue: NewIssuePage(debugFlag),
		},
	}
}

func (m *MaggiModel) Init() tea.Cmd {
	return nil
}

func (m *MaggiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for _, page := range m.pages {
			page.UpdateSize(msg.Width, msg.Height)
		}
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}
	// TODO: add current page invocation
	return nil, nil
}

func (m *MaggiModel) View() string {
	page, ok := m.pages[m.currentPage]
	if !ok {
		return "something horribly wrong with the app"
	}
	return page.View()
}
