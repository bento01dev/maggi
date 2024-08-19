package tui

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/bento01dev/maggi/internal/data"
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
	profiles []data.Profile
	err      error
}

type profileAddMsg struct {
	success bool
}

type profileDeleteMsg struct{}

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

// NOTE: new profile and update profile flow have a lot of overlap, but also checks and method calls
// can be slightly different. Code copying was simpler than trying to DRY it up too much.
// will revisit later if needed
type profileStage int

const (
	profileStageDefault profileStage = iota
	chooseAction
	addProfileName
	addProfileConfirm
	addProfileCancel
	updateProfileName
	updateProfileConfirm
	updateProfileCancel
	deleteProfileView
	deleteProfileConfirm
	deleteProfileCancel
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
	id     int
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

type profileModel interface {
	GetAll() ([]data.Profile, error)
	Add(name string) (data.Profile, error)
	Update(profile data.Profile, newName string) (data.Profile, error)
	Delete(profile data.Profile) error
}

type ProfilePage struct {
	newProfileOption  bool
	infoFlag          bool
	isErrInfo         bool
	width             int
	height            int
	currentUserFlow   profileUserFlow
	activePane        profilePagePane
	currentStage      profileStage
	profileModel      profileModel
	currentProfile    *data.Profile
	infoMsg           string
	actions           []string
	profiles          []data.Profile
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
	deleteButton      lipgloss.Style
	issuesStyle       lipgloss.Style
}

func NewProfilePage(profileDataModel *data.Profiles) *ProfilePage {
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
	actionsStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultActionsWidth).UnsetPadding()
	profilesStyle := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultProfileWidth).UnsetPadding()
	issuesStyle := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).UnsetPadding().BorderForeground(red)
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
	deleteButton := baseButton.Copy().Background(red)

	return &ProfilePage{
		profileModel:      profileDataModel,
		helpMenu:          helpMenu,
		keys:              keys,
		actionsStyle:      actionsStyle,
		profilesStyle:     profilesStyle,
		actions:           actions,
		titleStyle:        titleStyle,
		headingStyle:      headingStyle,
		issuesStyle:       issuesStyle,
		textInput:         input,
		highlightedButton: confirmButton,
		mutedButton:       cancelButton,
		deleteButton:      deleteButton,
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
		return p, func() tea.Msg {
			return p.getProfiles()
		}
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
		if msg.err != nil {
			return p, func() tea.Msg {
				return IssueMsg{Inner: msg.err}
			}
		}
		p.currentUserFlow = listProfiles
		p.profiles = msg.profiles
		p.activePane = profilesPane
		p.setActionsList()
		p.setProfileList()
		return p, nil
	case profileAddMsg, profileDeleteMsg:
		p.setActionsList()
		p.setProfileList()
		return p, nil
	}
	cmd := p.handleEvent(msg)
	return p, cmd
}

func (p *ProfilePage) resetInfoBag() {
	p.infoFlag = false
	p.isErrInfo = false
	p.infoMsg = ""
}

func (p *ProfilePage) getProfiles() retrieveMsg {
	profiles, err := p.profileModel.GetAll()
	if err != nil {
		return retrieveMsg{err: err}
	}
	return retrieveMsg{profiles: profiles}
}

// TODO: switch to sqlite
func (p *ProfilePage) addProfile(name string) error {
	profile, err := p.profileModel.Add(name)
	if err != nil {
		return err
	}
	p.profiles = append(p.profiles, profile)
	return nil
}

func (p *ProfilePage) updateProfile(profile *data.Profile, newName string) error {
    _, err := p.profileModel.Update(*profile, newName)
    if err != nil {
        return err
    }
	// var pos = -1
	// for i, profile := range p.profiles {
	// 	if profile == oldName {
	// 		pos = i
	// 		break
	// 	}
	// }
	// if pos == -1 {
	// 	return errors.New("profile not found")
	// }
	// p.profiles[pos] = newName
	return nil
}

func (p *ProfilePage) deleteProfile(profile *data.Profile) error {
    err := p.profileModel.Delete(*profile)
    if err != nil {
        return err 
    }
	// var pos int
	// for i, profile := range p.profiles {
	// 	if profile == name {
	// 		pos = i
	// 		break
	// 	}
	// }
	// p.profiles = append(p.profiles[:pos], p.profiles[pos+1:]...)
	return nil
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

func (p *ProfilePage) checkDuplicate(name string) bool {
	var names []string
	for _, profile := range p.profiles {
		names = append(names, profile.Name)
	}
	return slices.Contains(names, name)
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
    var err error 
    p.profiles, err = p.profileModel.GetAll()
    //TODO: it just returns for now. but need to exit with error 
    if err != nil {
        return
    }
	for _, profile := range p.profiles {
        profilesList = append(profilesList, profileItem{id: profile.ID, name: profile.Name})
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
		p.handleUpdateProfileTab(shift)
	case deleteProfile:
		p.handleDeleteProfileTab(shift)
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
				p.resetInfoBag()
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
			p.resetInfoBag()
		}
	}
}

func (p *ProfilePage) handleUpdateProfileTab(shift bool) {
	if shift {
		switch p.activePane {
		case profilesPane:
			p.activePane = actionsPane
		case actionsPane:
			switch p.currentStage {
			case updateProfileName:
				p.activePane = profilesPane
				p.resetInfoBag()
			case updateProfileConfirm:
				p.currentStage = updateProfileName
			case updateProfileCancel:
				p.currentStage = updateProfileConfirm
			}
		}
		return
	}

	switch p.activePane {
	case profilesPane:
		p.activePane = actionsPane
	case actionsPane:
		switch p.currentStage {
		case updateProfileName:
			p.currentStage = updateProfileConfirm
		case updateProfileConfirm:
			p.currentStage = updateProfileCancel
		case updateProfileCancel:
			p.currentStage = updateProfileName
			p.activePane = profilesPane
			p.resetInfoBag()
		}
	}
}

func (p *ProfilePage) handleDeleteProfileTab(shift bool) {
	if shift {
		switch p.activePane {
		case profilesPane:
			p.activePane = actionsPane
		case actionsPane:
			switch p.currentStage {
			case deleteProfileView:
				p.activePane = profilesPane
			case deleteProfileConfirm:
				p.currentStage = deleteProfileView
			case deleteProfileCancel:
				p.currentStage = deleteProfileConfirm
			}
		}
		return
	}

	switch p.activePane {
	case profilesPane:
		p.activePane = actionsPane
	case actionsPane:
		switch p.currentStage {
		case deleteProfileView:
			p.currentStage = deleteProfileConfirm
		case deleteProfileConfirm:
			p.currentStage = deleteProfileCancel
		case deleteProfileCancel:
			p.currentStage = deleteProfileView
			p.activePane = profilesPane
		}
	}
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
		p.currentProfile = &data.Profile{ID: item.id, Name: item.name}
		p.currentStage = chooseAction
		return nil
	case actionsPane:
		item, ok := p.actionList.SelectedItem().(actionItem)
		if !ok {
			return func() tea.Msg {
				return errors.New("item not found in list. unknown issue")
			}
		}
		p.currentUserFlow = item.next
		switch p.currentUserFlow {
		case updateProfile:
			p.infoFlag = true
			p.infoMsg = fmt.Sprintf("You are trying to update %s with a new name. Please follow the instructions below.", p.currentProfile)
			p.currentStage = updateProfileName
			return tea.Batch(p.textInput.Focus(), p.textInput.Cursor.BlinkCmd())
		case deleteProfile:
			p.currentStage = deleteProfileView
		}
		return nil
	}
	return nil
}

func (p *ProfilePage) handleNewProfileEnter() tea.Cmd {
	p.resetInfoBag()
	switch p.currentStage {
	case addProfileName:
		if strings.TrimSpace(p.textInput.Value()) == "" {
			p.infoFlag = true
			p.isErrInfo = true
			p.infoMsg = "Please pass a valid profile name. You can exit flow by pressing <esc> if needed"
			return tea.Batch(p.textInput.Focus(), p.textInput.Cursor.BlinkCmd())
		}
		p.currentStage = addProfileConfirm
		return nil
	case addProfileConfirm:
		name := strings.TrimSpace(p.textInput.Value())
		if p.checkDuplicate(name) {
			p.infoMsg = fmt.Sprintf("The name %s is already taken. Try another name or update the existing one first!", name)
			p.infoFlag = true
			p.isErrInfo = true
			p.textInput.SetValue("")
			p.currentStage = addProfileName
			p.issuesStyle = p.issuesStyle.Copy().Width(len(p.infoMsg) + 1)
			return tea.Batch(p.textInput.Focus(), p.textInput.Cursor.BlinkCmd())
		}
		p.currentUserFlow = listProfiles
		p.currentStage = chooseAction
		p.activePane = profilesPane

		return func() tea.Msg {
			err := p.addProfile(name)
			if err != nil {
				return IssueMsg{Inner: err}
			}
			p.textInput.SetValue("")
			return profileAddMsg{success: true}
		}
	case addProfileCancel:
		p.currentStage = chooseAction
		p.textInput.SetValue("")
		p.currentUserFlow = listProfiles
		p.activePane = profilesPane
		p.updateActionStyle()
		p.updateProfileStyle()
		return nil
	}
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
	p.resetInfoBag()
	input := strings.TrimSpace(p.textInput.Value())
	switch p.currentStage {
	case updateProfileName:
		if input == "" || input == p.currentProfile.Name {
			p.infoFlag = true
			p.isErrInfo = true
			p.infoMsg = fmt.Sprintf("Please provide a valid new name to update %s. You can exit flow by pressing <esc> if needed", p.currentProfile)
			return tea.Batch(p.textInput.Focus(), p.textInput.Cursor.BlinkCmd())
		}
		p.currentStage = updateProfileConfirm
		return nil
	case updateProfileConfirm:
		if p.checkDuplicate(input) {
			p.infoMsg = fmt.Sprintf("The name %s is already taken. Try another name or update the existing one first!", input)
			p.infoFlag = true
			p.isErrInfo = true
			p.textInput.SetValue("")
			p.currentStage = updateProfileName
			p.issuesStyle = p.issuesStyle.Copy().Width(len(p.infoMsg) + 1)
			return tea.Batch(p.textInput.Focus(), p.textInput.Cursor.BlinkCmd())
		}
		p.currentUserFlow = listProfiles
		p.currentStage = chooseAction
		p.activePane = profilesPane

		return func() tea.Msg {
			err := p.updateProfile(p.currentProfile, input)
			if err != nil {
				return IssueMsg{Inner: err}
			}
			p.textInput.SetValue("")
			return profileAddMsg{success: true}
		}
	case updateProfileCancel:
		p.currentStage = chooseAction
		p.textInput.SetValue("")
		p.currentUserFlow = listProfiles
		p.activePane = profilesPane
		p.updateActionStyle()
		p.updateProfileStyle()
		return nil
	}
	return nil
}

func (p *ProfilePage) handleDeleteProfileEnter() tea.Cmd {
	switch p.currentStage {
	case deleteProfileView:
		p.currentStage = deleteProfileConfirm
		return nil
	case deleteProfileCancel:
		p.currentStage = chooseAction
		p.currentProfile = nil
		p.activePane = profilesPane
		p.currentUserFlow = listProfiles
		p.updateActionStyle()
		p.updateProfileStyle()
		return nil
	case deleteProfileConfirm:
		p.currentUserFlow = listProfiles
		p.currentStage = chooseAction
		p.activePane = profilesPane
		p.setProfileList()
		p.setActionsList()

		return func() tea.Msg {
			err := p.deleteProfile(p.currentProfile)
			if err != nil {
				return IssueMsg{Inner: err}
			}
			p.currentProfile = nil
			return profileDeleteMsg{}
		}
	}
	return nil
}

func (p *ProfilePage) handleEsc() {
	p.currentUserFlow = listProfiles
	p.activePane = profilesPane
	p.updateActionStyle()
	p.updateProfileStyle()
	p.textInput.SetValue("")
	p.infoMsg = ""
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
		case newProfile, updateProfile:
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
	case updateProfile:
		second = fmt.Sprintf(" Update Profile | %s ", p.currentProfile)
	case deleteProfile:
		second = fmt.Sprintf(" Delete Profile | %s ", p.currentProfile)
	}
	third := strings.Repeat("-", (defaultWidth - (len(second) + 3)))
	return first + second + third
}

func (p *ProfilePage) viewNewProfile() string {
	var textInputStyle, confirmButtonStyle, cancelButtonStyle, infoStyle lipgloss.Style

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
			confirmButtonStyle = p.mutedButton.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(blue)
		}
		cancelButtonStyle = p.mutedButton.Copy()
	case addProfileCancel:
		textInputStyle = p.actionsStyle.BorderForeground(muted)
		confirmButtonStyle = p.mutedButton.Copy()
		cancelButtonStyle = p.highlightedButton.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(blue)
	}

	if p.infoFlag {
		infoStyle = p.issuesStyle.Copy().BorderForeground(green)
		if p.isErrInfo {
			infoStyle = p.issuesStyle.Copy().BorderForeground(red)
		}

		return lipgloss.Place(
			p.width,
			p.height,
			lipgloss.Center,
			lipgloss.Center,
			lipgloss.JoinVertical(
				lipgloss.Center,
				p.titleStyle.Render(p.generateTitle()),
				infoStyle.Render(p.infoMsg),
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

func (p *ProfilePage) viewUpdateProfile() string {
	var textInputStyle, confirmButtonStyle, cancelButtonStyle, infoStyle lipgloss.Style
	switch p.currentStage {
	case updateProfileName:
		textInputStyle = p.actionsStyle.Copy()
		confirmButtonStyle = p.mutedButton.Copy()
		if p.textInput.Value() != "" {
			confirmButtonStyle = p.highlightedButton.Copy()
		}
		cancelButtonStyle = p.mutedButton.Copy()
	case updateProfileConfirm:
		textInputStyle = p.actionsStyle.BorderForeground(muted)
		confirmButtonStyle = p.highlightedButton.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(blue)
		if p.textInput.Value() == "" {
			confirmButtonStyle = p.mutedButton.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(blue)
		}
		cancelButtonStyle = p.mutedButton.Copy()
	case updateProfileCancel:
		textInputStyle = p.actionsStyle.BorderForeground(muted)
		confirmButtonStyle = p.mutedButton.Copy()
		cancelButtonStyle = p.highlightedButton.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(blue)
	}

	if p.infoFlag {
		infoStyle = p.issuesStyle.Copy().BorderForeground(green)
		if p.isErrInfo {
			infoStyle = p.issuesStyle.Copy().BorderForeground(red)
		}

		return lipgloss.Place(
			p.width,
			p.height,
			lipgloss.Center,
			lipgloss.Center,
			lipgloss.JoinVertical(
				lipgloss.Center,
				p.titleStyle.Render(p.generateTitle()),
				infoStyle.Render(p.infoMsg),
				lipgloss.JoinHorizontal(
					lipgloss.Center,
					p.profilesStyle.Render(p.profileList.View()),
					lipgloss.JoinVertical(
						lipgloss.Left,
						p.headingStyle.Render("New Name:"),
						textInputStyle.Render(p.textInput.View()),
						lipgloss.JoinHorizontal(
							lipgloss.Center,
							confirmButtonStyle.Render("Update"),
							cancelButtonStyle.Render("Cancel"),
						),
					),
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
				lipgloss.JoinVertical(
					lipgloss.Left,
					p.headingStyle.Render("New Name:"),
					textInputStyle.Render(p.textInput.View()),
					lipgloss.JoinHorizontal(
						lipgloss.Center,
						confirmButtonStyle.Render("Update"),
						cancelButtonStyle.Render("Cancel"),
					),
				),
			),
			p.helpMenu.View(p.keys),
		),
	)
}

func (p *ProfilePage) viewListProfile() string {
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
}

func (p *ProfilePage) viewDeleteProfile() string {
	msg := fmt.Sprintf("Deleting profile %s will also delete all the aliases and envs attached to the profile. Are you sure?", p.currentProfile.Name)
	paddingTotal := defaultActionsWidth - len(msg)
	var infoStyle, deleteButton, cancelButton lipgloss.Style
	infoStyle = p.issuesStyle.Copy().BorderForeground(red).PaddingLeft(paddingTotal / 2).PaddingRight(paddingTotal / 2)
	switch p.currentStage {
	case deleteProfileView:
		deleteButton = p.deleteButton
		cancelButton = p.mutedButton
	case deleteProfileConfirm:
		deleteButton = p.deleteButton.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(red)
		cancelButton = p.mutedButton
		infoStyle = infoStyle.Copy().BorderForeground(muted)
	case deleteProfileCancel:
		deleteButton = p.deleteButton
		cancelButton = p.highlightedButton.Copy().Border(lipgloss.DoubleBorder()).BorderForeground(green)
		infoStyle = infoStyle.Copy().BorderForeground(muted)
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
					lipgloss.Center,
					infoStyle.Render(msg),
					lipgloss.JoinHorizontal(
						lipgloss.Center,
						deleteButton.Render("Yes, Delete"),
						cancelButton.Render("Cancel"),
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
		return p.viewListProfile()
	case newProfile:
		return p.viewNewProfile()
	case updateProfile:
		return p.viewUpdateProfile()
	case deleteProfile:
		return p.viewDeleteProfile()
	}
	return "profile page.."
}

func (p *ProfilePage) UpdateSize(width, height int) {
	p.width = width
	p.height = height
}
