package tui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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

const (
	buttonPaddingHorizontal int = 2
	buttonPaddingVertical   int = 0
)

type profileUserFlow int

const (
	retrieveProfiles profileUserFlow = iota
	listProfiles
	newProfile
	viewProfile
	updateProfile
	deleteProfile
)

type profilePagePane int

const (
	profilesPane profilePagePane = iota
	actionsPane
)

type profileStage int

const (
	profileStageDefault profileStage = iota
	chooseAction
	addProfileName
	addProfileConfirm
	addProfileCancel
	updateProfileName
	updateProfileConfirm
	deleteProfileConfirm
)

type actionItem struct {
	next        profileUserFlow
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
	newProfileOption  bool
	width             int
	height            int
	currentUserFlow   profileUserFlow
	activePane        profilePagePane
	currentStage      profileStage
	currentProfile    string
	actions           []string
	profiles          []string
	actionList        list.Model
	actionsStyle      lipgloss.Style
	profileList       list.Model
	profilesStyle     lipgloss.Style
	helpMenu          help.Model
	keys              profileHelpKeys
	titleStyle        lipgloss.Style
	headingStyle      lipgloss.Style
	textInput         textinput.Model
	highlightedButton lipgloss.Style
	mutedButton       lipgloss.Style
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
	titleStyle := lipgloss.NewStyle().Foreground(green)
	headingStyle := lipgloss.NewStyle().Foreground(blue)

	input := textinput.New()
	input.Placeholder = "Your new profile name.."
	input.CharLimit = 50
	input.Width = 50
	input.Prompt = ""

	baseButton := lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0"))
	confirmButton := baseButton.Copy().Background(green)
	cancelButton := baseButton.Copy().Background(muted)

	return &ProfilePage{
		helpMenu:          helpMenu,
		keys:              keys,
		actionsStyle:      actionsStyle,
		profilesStyle:     profilesStyle,
		actions:           actions,
		titleStyle:        titleStyle,
		headingStyle:      headingStyle,
		textInput:         input,
		highlightedButton: confirmButton,
		mutedButton:       cancelButton,
	}
}

func (p *ProfilePage) Init() tea.Cmd {
	return nil
}

//TODO: handle enter. do for each stage and pane combo
//TODO: figure out height
//TODO: default list when no profiles

func (p *ProfilePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProfileStartMsg:
		p.currentUserFlow = retrieveProfiles
		return p, p.getProfiles()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			p.handleTab(false)
			return p, nil
		case tea.KeyShiftTab:
			p.handleTab(true)
			return p, nil
		case tea.KeyEnter:
			return p, p.handleEnter()
		case tea.KeyEsc:
			p.handleEsc()
			return p, nil
		}
	case retrieveMsg:
		p.currentUserFlow = listProfiles
		p.profiles = msg.profiles
		p.activePane = profilesPane
		p.setActionsList()
		p.setProfileList()
		return p, nil
	}
	cmd := p.handleEvent(msg)
	return p, cmd
}

func (p *ProfilePage) getProfiles() tea.Cmd {
	return func() tea.Msg {
		return retrieveMsg{profiles: []string{"global", "stg", "prd", "dev"}}
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
	case profilesPane:
		p.actionsStyle = p.actionsStyle.Copy().BorderForeground(muted)
	case actionsPane:
		p.actionsStyle = p.actionsStyle.Copy().BorderForeground(green)
	}
}

func (p *ProfilePage) updateProfileStyle() {
	switch p.activePane {
	case profilesPane:
		p.profilesStyle = p.profilesStyle.Copy().BorderForeground(green)
	case actionsPane:
		p.profilesStyle = p.profilesStyle.Copy().BorderForeground(muted)
	}
}

func (p *ProfilePage) handleTab(shift bool) {
	switch p.currentUserFlow {
	case listProfiles:
		p.handleListProfilesTab()
	case newProfile:
		p.handleNewProfileTab(shift)
	case updateProfile:
		p.handleUpdateProfileTab()
	case deleteProfile:
		p.handleDeleteProfileTab()
	}
	p.updateActionStyle()
	p.updateProfileStyle()
}

func (p *ProfilePage) handleListProfilesTab() {
	switch p.activePane {
	case profilesPane:
		p.activePane = actionsPane
	case actionsPane:
		p.activePane = profilesPane
	}
}

func (p *ProfilePage) handleNewProfileTab(shift bool) {
	if shift {
		switch p.activePane {
		case profilesPane:
			p.activePane = actionsPane
		case actionsPane:
			switch p.currentStage {
			case addProfileName:
				p.activePane = profilesPane
			case addProfileConfirm:
				p.currentStage = addProfileName
			case addProfileCancel:
				p.currentStage = addProfileConfirm
			}
		}
		return
	}
	switch p.activePane {
	case profilesPane:
		p.activePane = actionsPane
	case actionsPane:
		switch p.currentStage {
		case addProfileName:
			p.currentStage = addProfileConfirm
		case addProfileConfirm:
			p.currentStage = addProfileCancel
		case addProfileCancel:
			p.currentStage = addProfileName
			p.activePane = profilesPane
		}
	}
}

func (p *ProfilePage) handleUpdateProfileTab() {
}

func (p *ProfilePage) handleDeleteProfileTab() {
}

func (p *ProfilePage) handleEnter() tea.Cmd {
	switch p.currentUserFlow {
	case listProfiles:
		return p.handleListProfilesEnter()
	case newProfile:
		return p.handleNewProfileEnter()
	case viewProfile:
		return p.handleViewProfileEnter()
	case updateProfile:
		return p.handleUpdateProfileEnter()
	case deleteProfile:
		return p.handleDeleteProfileEnter()
	default:
		return nil
	}
}

func (p *ProfilePage) handleListProfilesEnter() tea.Cmd {
	switch p.activePane {
	case profilesPane:
		item, ok := p.profileList.SelectedItem().(profileItem)
		if !ok {
			return func() tea.Msg {
				return errors.New("item not found in list. unknown issue")
			}
		}
		p.activePane = actionsPane
		p.updateActionStyle()
		p.updateProfileStyle()
		if item.action {
			p.currentUserFlow = newProfile
			p.currentStage = addProfileName
			return tea.Batch(p.textInput.Focus(), p.textInput.Cursor.BlinkCmd())
		}
		p.currentProfile = item.name
		return nil
	}
	return nil
}

func (p *ProfilePage) handleNewProfileEnter() tea.Cmd {
	return nil
}

func (p *ProfilePage) handleViewProfileEnter() tea.Cmd {
	return func() tea.Msg {
		item, ok := p.profileList.SelectedItem().(profileItem)
		if !ok {
			return errors.New("unknown item in list")
		}
		return ProfileDoneMsg{profile: item.name}
	}
}

func (p *ProfilePage) handleUpdateProfileEnter() tea.Cmd {
	return nil
}

func (p *ProfilePage) handleDeleteProfileEnter() tea.Cmd {
	return nil
}

func (p *ProfilePage) handleEsc() {
	p.currentUserFlow = listProfiles
	p.activePane = profilesPane
	p.updateActionStyle()
	p.updateProfileStyle()
	p.textInput.SetValue("")
}

func (p *ProfilePage) handleEvent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch p.activePane {
	case profilesPane:
		p.profileList, cmd = p.profileList.Update(msg)
		item, ok := p.profileList.SelectedItem().(profileItem)
		if !ok {
			return tea.Quit
		}
		if item.action {
			p.newProfileOption = true
		} else {
			p.newProfileOption = false
			p.currentUserFlow = listProfiles
		}
	case actionsPane:
		switch p.currentUserFlow {
		case listProfiles:
			p.actionList, cmd = p.actionList.Update(msg)
		case newProfile:
			p.textInput, cmd = p.textInput.Update(msg)
		}
	}
	return cmd
}

func (p *ProfilePage) generateTitle() string {
	first := strings.Repeat("-", 3)
	var second string
	switch p.currentUserFlow {
	case listProfiles:
		second = " Profiles "
	case newProfile:
		second = " New Profile "
	}
	third := strings.Repeat("-", (defaultWidth - (len(second) + 3)))
	return first + second + third
}

func (p *ProfilePage) viewNewProfile() string {
	var textInputStyle, confirmButtonStyle, cancelButtonStyle lipgloss.Style

	switch p.currentStage {
	case addProfileName:
		textInputStyle = p.actionsStyle.Copy()
		confirmButtonStyle = p.mutedButton.Copy()
		if p.textInput.Value() != "" {
			confirmButtonStyle = p.highlightedButton.Copy()
		}
		cancelButtonStyle = p.mutedButton.Copy()
	case addProfileConfirm:
		textInputStyle = p.actionsStyle.BorderForeground(muted)
		confirmButtonStyle = p.highlightedButton.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(blue)
		if p.textInput.Value() == "" {
			confirmButtonStyle = p.mutedButton.Copy()
		}
		cancelButtonStyle = p.mutedButton.Copy()
	case addProfileCancel:
		textInputStyle = p.actionsStyle.BorderForeground(muted)
		confirmButtonStyle = p.mutedButton.Copy()
		cancelButtonStyle = p.highlightedButton.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(blue)
	}

	return lipgloss.Place(
		p.width,
		p.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			p.titleStyle.Render(p.generateTitle()),
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				p.profilesStyle.Render(p.profileList.View()),
				lipgloss.JoinVertical(
					lipgloss.Left,
					p.headingStyle.Render("Name:"),
					textInputStyle.Render(p.textInput.View()),
					lipgloss.JoinHorizontal(
						lipgloss.Center,
						confirmButtonStyle.Render("Create"),
						cancelButtonStyle.Render("Cancel"),
					),
				),
			),
			p.helpMenu.View(p.keys),
		),
	)
}

func (p *ProfilePage) View() string {
	switch p.currentUserFlow {
	case listProfiles:
		if p.newProfileOption {
			return lipgloss.Place(
				p.width,
				p.height,
				lipgloss.Center,
				lipgloss.Center,
				lipgloss.JoinVertical(
					lipgloss.Center,
					p.titleStyle.Render(p.generateTitle()),
					lipgloss.JoinHorizontal(
						lipgloss.Center,
						p.profilesStyle.Render(p.profileList.View()),
						p.actionsStyle.Copy().Height(len(p.profiles)+1).Render(""),
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
				p.titleStyle.Render(p.generateTitle()),
				lipgloss.JoinHorizontal(
					lipgloss.Center,
					p.profilesStyle.Render(p.profileList.View()),
					p.actionsStyle.Render(p.actionList.View()),
				),
				p.helpMenu.View(p.keys),
			),
		)
	case newProfile:
		return p.viewNewProfile()
	}
	return "profile page.."
}

func (p *ProfilePage) UpdateSize(width, height int) {
	p.width = width
	p.height = height
}
