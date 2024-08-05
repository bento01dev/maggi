package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProfileStartMsg struct{}

type ProfileDoneMsg struct {
	profile string
}

func (p ProfileDoneMsg) Next() pageType {
	return home
}

type retrieveMsg struct {
	profiles []string
}

type profileStage int

const (
	retrieveProfiles profileStage = iota
	listProfiles
	newProfile
	updateProfile
	deleteProfile
)

type actionItem struct {
	next        profileStage
	description string
}

func (a actionItem) FilterValue() string { return "" }
func renderActionItem(i list.Item) string {
	a, ok := i.(actionItem)
	if !ok {
		return ""
	}
	return a.description
}

type profileItem struct {
	name string
}

func (p profileItem) FilterValue() string { return "" }
func renderProfileItem(i list.Item) string {
	p, ok := i.(profileItem)
	if !ok {
		return ""
	}
	return p.name
}

type profileHelpKeys struct {
	ToggleView key.Binding
	Quit       key.Binding
	Up         key.Binding
	Down       key.Binding
	Esc        key.Binding
}

func (h profileHelpKeys) ShortHelp() []key.Binding {
	return []key.Binding{h.ToggleView, h.Up, h.Down, h.Esc, h.Quit}
}

func (h profileHelpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{h.ToggleView, h.Up, h.Down},
		{h.Esc, h.Quit},
	}
}

type ProfilePage struct {
	width         int
	height        int
	currentStage  profileStage
	actions       list.Model
	actionsStyle  lipgloss.Style
	profiles      list.Model
	profilesStyle lipgloss.Style
	helpMenu      help.Model
	keys          profileHelpKeys
}

func NewProfilePage() *ProfilePage {
	helpMenu := help.New()
	keyStyle := lipgloss.NewStyle().Foreground(muted)
	descStyle := lipgloss.NewStyle().Foreground(muted)
	sepStyle := lipgloss.NewStyle().Foreground(muted)
	helpStyles := help.Styles{
		ShortKey:       keyStyle,
		ShortDesc:      descStyle,
		ShortSeparator: sepStyle,
		Ellipsis:       sepStyle.Copy(),
		FullKey:        keyStyle.Copy(),
		FullDesc:       descStyle.Copy(),
		FullSeparator:  sepStyle.Copy(),
	}
	helpMenu.Styles = helpStyles
	keys := profileHelpKeys{
		ToggleView: key.NewBinding(
			key.WithKeys("<tab>"),
			key.WithHelp("<tab>", "toggle panes"),
		),
		Quit: key.NewBinding(
			key.WithKeys("<ctrl+c>"),
			key.WithHelp("<ctrl+c>", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "move down"),
		),
		Esc: key.NewBinding(
			key.WithKeys("<esc>"),
			key.WithHelp("<esc>", "quit view"),
		),
	}
	actionsList := []list.Item{
		actionItem{
			description: "Add Profile",
			next:        newProfile,
		},
		actionItem{
			description: "Delete Profile",
			next:        deleteProfile,
		},
		actionItem{
			description: "Update Profile",
			next:        updateProfile,
		},
	}
	actions := list.New(actionsList, ListItemDelegate{RenderFunc: renderActionItem}, 15, 2)
	actions.SetShowTitle(false)
	actions.SetShowStatusBar(false)
	actions.SetShowFilter(false)
	actions.SetFilteringEnabled(false)
	actions.SetShowPagination(false)
	actions.SetShowHelp(false)

	actionsStyle := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(green).Width(20).UnsetPadding()
	profilesStyle := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(green).Width(20).UnsetPadding()

	return &ProfilePage{
		helpMenu:      helpMenu,
		keys:          keys,
		actions:       actions,
		actionsStyle:  actionsStyle,
		profilesStyle: profilesStyle,
	}
}

func (p *ProfilePage) Init() tea.Cmd {
	return nil
}

func (p *ProfilePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProfileStartMsg:
		p.currentStage = newProfile
		return p, p.getProfiles()
	case retrieveMsg:
		p.currentStage = listProfiles
		p.setProfileList(msg.profiles)
		return p, nil
	}
	return nil, nil
}

func (p *ProfilePage) getProfiles() tea.Cmd {
	return func() tea.Msg {
		return retrieveMsg{profiles: []string{"global", "stg", "prd"}}
	}
}

func (p *ProfilePage) setProfileList(profileStrs []string) {
	profilesList := []list.Item{}
	for _, profileStr := range profileStrs {
		profilesList = append(profilesList, profileItem{name: profileStr})
	}
	profiles := list.New(profilesList, ListItemDelegate{RenderFunc: renderProfileItem}, 15, 2)
	profiles.SetShowTitle(false)
	profiles.SetShowStatusBar(false)
	profiles.SetShowFilter(false)
	profiles.SetFilteringEnabled(false)
	profiles.SetShowPagination(false)
	profiles.SetShowHelp(false)

	p.profiles = profiles
}

func (p *ProfilePage) View() string {
	switch p.currentStage {
	case listProfiles:
		return lipgloss.Place(
			p.width,
			p.height,
			lipgloss.Center,
			lipgloss.Center,
			lipgloss.JoinVertical(
				lipgloss.Center,
				lipgloss.JoinHorizontal(
					lipgloss.Center,
					p.profilesStyle.Render(p.profiles.View()),
					p.actionsStyle.Render(p.actions.View()),
				),
				p.helpMenu.View(p.keys),
			),
		)
	}
	return lipgloss.Place(
		p.width,
		p.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			p.actionsStyle.Render(p.actions.View()),
			p.helpMenu.View(p.keys),
		),
	)
}

func (p *ProfilePage) UpdateSize(width, height int) {
	p.width = width
	p.height = height
}
