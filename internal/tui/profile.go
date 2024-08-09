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

const (
	defaultWidth        int = 120
	defaultProfileWidth int = 30
	defaultActionsWidth int = 90
)

type profileStage int

const (
	retrieveProfiles profileStage = iota
	listProfiles
	newProfile
	viewProfile
	updateProfile
	deleteProfile
)

type profilePane int

const (
	profilesListPane profilePane = iota
	actionsListPane
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
	name   string
	action bool
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
	activePane    profilePane
	actions       []string
	profiles      []string
	actionList    list.Model
	actionsStyle  lipgloss.Style
	profileList   list.Model
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
	actionsStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(90).UnsetPadding()
	profilesStyle := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(30).UnsetPadding()
	actions := []string{"View Profile", "Update Profile", "Delete Profile"}

	return &ProfilePage{
		helpMenu:      helpMenu,
		keys:          keys,
		actionsStyle:  actionsStyle,
		profilesStyle: profilesStyle,
		actions:       actions,
	}
}

func (p *ProfilePage) Init() tea.Cmd {
	return nil
}

func (p *ProfilePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProfileStartMsg:
		p.currentStage = retrieveProfiles
		return p, p.getProfiles()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			p.handleTab()
			return p, nil
		}
	case retrieveMsg:
		p.currentStage = listProfiles
		p.profiles = msg.profiles
		p.activePane = profilesListPane
		p.setActionsList()
		p.setProfileList()
		return p, nil
	}
	cmd := p.handleEvent(msg)
	return p, cmd
}

func (p *ProfilePage) getProfiles() tea.Cmd {
	return func() tea.Msg {
		return retrieveMsg{profiles: []string{"global", "stg", "prd"}}
	}
}

func (p *ProfilePage) getItemsMaxLen(elems []string) int {
	var width int
	for _, elem := range elems {
		if len(elem) >= width {
			width = len(elem)
		}
	}
	return width
}

func (p *ProfilePage) setActionsList() {
	actionItems := []actionItem{
		actionItem{
			description: "View Profile",
			next:        viewProfile,
		},
		actionItem{
			description: "Update Profile",
			next:        updateProfile,
		},
		actionItem{
			description: "Delete Profile",
			next:        deleteProfile,
		},
	}
	var elems []string
	var actionsList []list.Item
	for _, a := range actionItems {
		elems = append(elems, a.description)
		actionsList = append(actionsList, a)
	}
	w := defaultActionsWidth
	if p.width < defaultWidth {
		w = p.width - defaultProfileWidth
	}

	h := len(actionsList)
	if (len(p.profiles) + 1) > h {
		h = len(p.profiles) + 1
	}

	p.actionList = GenerateList(actionsList, renderActionItem, w, h)
	p.updateActionStyle()
}

func (p *ProfilePage) setProfileList() {
	profilesList := []list.Item{}
	for _, profileStr := range p.profiles {
		profilesList = append(profilesList, profileItem{name: profileStr})
	}
	profilesList = append(profilesList, profileItem{name: "Add Profile...", action: true})
	p.profileList = GenerateList(profilesList, renderProfileItem, 30, len(profilesList))
	p.updateProfileStyle()
}

func (p *ProfilePage) updateActionStyle() {
	switch p.activePane {
	case profilesListPane:
		p.actionsStyle = p.actionsStyle.Copy().BorderForeground(muted)
	case actionsListPane:
		p.actionsStyle = p.actionsStyle.Copy().BorderForeground(green)
	}
}

func (p *ProfilePage) updateProfileStyle() {
	switch p.activePane {
	case profilesListPane:
		p.profilesStyle = p.profilesStyle.Copy().BorderForeground(green)
	case actionsListPane:
		p.profilesStyle = p.profilesStyle.Copy().BorderForeground(muted)
	}
}

func (p *ProfilePage) handleTab() {
	switch p.activePane {
	case profilesListPane:
		p.activePane = actionsListPane
	case actionsListPane:
		p.activePane = profilesListPane
	}
	p.updateActionStyle()
	p.updateProfileStyle()
}

func (p *ProfilePage) handleEvent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch p.activePane {
	case profilesListPane:
		p.profileList, cmd = p.profileList.Update(msg)
	case actionsListPane:
		p.actionList, cmd = p.actionList.Update(msg)
	}
	return cmd
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
					p.profilesStyle.Render(p.profileList.View()),
					p.actionsStyle.Render(p.actionList.View()),
				),
				p.helpMenu.View(p.keys),
			),
		)
	}
	return "profile page.."
}

func (p *ProfilePage) UpdateSize(width, height int) {
	p.width = width
	p.height = height
}
