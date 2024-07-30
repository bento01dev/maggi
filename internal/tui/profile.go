package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ProfileStartMsg struct{}

type ProfileDoneMsg struct {
	profileId int
}

func (p ProfileDoneMsg) Next() pageType {
	return home
}

type profileStage int

const (
	listProfiles profileStage = iota
	newProfile
	updateProfile
	deleteProfile
)

type helpKeys struct {
	AltView key.Binding
	Quit    key.Binding
	Up      key.Binding
	Down    key.Binding
	Esc     key.Binding
}

func (h helpKeys) ShortHelp() []key.Binding {
	return []key.Binding{h.AltView, h.Up, h.Down, h.Esc, h.Quit}
}

func (h helpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{h.AltView, h.Up, h.Down},
		{h.Esc, h.Quit},
	}
}

type ProfilePage struct {
	width    int
	height   int
	actions  list.Model
	profiles list.Model
	helpMenu help.Model
	keys     helpKeys
}

func NewProfilePage() *ProfilePage {
	return &ProfilePage{}
}

func (p *ProfilePage) Init() tea.Cmd {
	return nil
}

func (p *ProfilePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (p *ProfilePage) View() string {
	return "profile page"
}

func (p *ProfilePage) UpdateSize(width, height int) {
	p.width = width
	p.height = height
}
