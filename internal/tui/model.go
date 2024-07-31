package tui

import (
	"errors"

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
	start pageType = iota
	profile
	home
	issue
)

type Page interface {
	tea.Model
	UpdateSize(width, height int)
}

type PageTurner interface {
	Next() pageType
}

type GenericTurner pageType

func (g GenericTurner) Next() pageType {
	return pageType(g)
}

type MaggiModel struct {
	currentPage pageType
	pages       map[pageType]Page
	profile     string
	err         error
}

func NewMaggiModel(debugFlag bool) *MaggiModel {
	return &MaggiModel{
		pages: map[pageType]Page{
			issue:   NewIssuePage(debugFlag),
			profile: NewProfilePage(),
		},
	}
}

func (m *MaggiModel) Init() tea.Cmd {
	m.currentPage = start
	return func() tea.Msg {
		return GenericTurner(profile)
	}
}

func (m *MaggiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for _, page := range m.pages {
			page.UpdateSize(msg.Width, msg.Height)
		}
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case PageTurner:
		m.handlePageTransition(msg)
		return m, m.pageInitCmd()
	}
	page, ok := m.pages[m.currentPage]
	if !ok {
		m.err = errors.New("unknown page lookup")
		m.currentPage = issue
		return m, nil
	}
	_, cmd = page.Update(msg)
	return m, cmd
}

func (m *MaggiModel) handlePageTransition(msg PageTurner) {
	switch msg := msg.(type) {
	case ProfileDoneMsg:
		m.profile = msg.profile
	}
	m.currentPage = msg.Next()
}

func (m *MaggiModel) pageInitCmd() tea.Cmd {
	var msg tea.Msg
	switch m.currentPage {
	case start:
		msg = ProfileStartMsg{}
	}
	return func() tea.Msg {
		return msg
	}
}

func (m *MaggiModel) View() string {
	page, ok := m.pages[m.currentPage]
	if !ok {
		return "something horribly wrong with the app"
	}
	return page.View()
}
