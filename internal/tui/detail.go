package tui

import (
	"github.com/bento01dev/maggi/internal/data"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DetailStartMsg struct {
	currentProfile data.Profile
}

type retrieveDetailsMsg struct {
	details []data.Detail
	err     error
}

const (
	defaultDPWidth      int = 120
	defaultSideBarWidth int = 30
	defaultDisplayWidth int = 90
	defaultDPHeight     int = 20
)

type detailsUserFlow int

const (
	retrieveDetails detailsUserFlow = iota
	listDetails
	viewDetail
	newDetail
	updateDetail
	deleteDetail
)

type detailPagePane int

const (
	envPane detailPagePane = iota
	aliasPane
	detailActionPane
)

type detailType int

const (
	detailTypeDefault detailType = iota
	detailTypeEnv
	detailTypeAlias
)

type detailStage int

const (
	detailStageDefault detailStage = iota
	chooseDetailAction
	addDetailName
	addDetailConfirm
	addDetailCancel
	updateDetailName
	updateDetailConfirm
	updateDetailCancel
	deleteDetailView
	deleteDetailConfirm
	deleteDetailCancel
)

type detailHelpKeys struct {
	ToggleView key.Binding
	Quit       key.Binding
	Up         key.Binding
	Down       key.Binding
	Esc        key.Binding
}

func (h detailHelpKeys) ShortHelp() []key.Binding {
	return []key.Binding{h.ToggleView, h.Up, h.Down, h.Esc, h.Quit}
}

func (h detailHelpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{h.ToggleView, h.Up, h.Down},
		{h.Esc, h.Quit},
	}
}

type detailModel interface {
	GetAll(profileId int) ([]data.Detail, error)
	Add(key string, value string, detailType data.DetailType, profileID int) (data.Detail, error)
	Update(detail data.Detail, key string, value string) (data.Detail, error)
	Delete(detail data.Detail) error
	DeleteAll(profileID int) error
}

type DetailPage struct {
	width             int
	height            int
	currentUserFlow   detailsUserFlow
	detailType        detailType
	currentStage      detailStage
	activePane        detailPagePane
	currentProfile    data.Profile
	detailModel       detailModel
	details           []data.Detail
	helpMenu          help.Model
	keys              detailHelpKeys
	titleStyle        lipgloss.Style
	headingStyle      lipgloss.Style
	issuesStyle       lipgloss.Style
	actionsStyle      lipgloss.Style
	detailsStyle      lipgloss.Style
	aliasList         list.Model
	envList           list.Model
	keyInput          textinput.Model
	valueInput        textinput.Model
	highlightedButton lipgloss.Style
	mutedButton       lipgloss.Style
	deleteButton      lipgloss.Style
	infoMsg           string
	actions           []string
}

//TODO: the styling elements and help menu are duplicates of what is in profile. de-duplicate if needed later.
func NewDetailPage(detailDataModel detailModel) *DetailPage {
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
	keys := detailHelpKeys{
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
	detailStyle := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultProfileWidth).UnsetPadding()
	issuesStyle := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).UnsetPadding().BorderForeground(red)
	actions := []string{"Update Detail", "Delete Detail"}
	titleStyle := lipgloss.NewStyle().Foreground(green)
	headingStyle := lipgloss.NewStyle().Foreground(blue)

	keyInput := textinput.New()
	keyInput.Placeholder = "Name.."
	keyInput.CharLimit = 50
	keyInput.Width = 50
	keyInput.Prompt = ""

	valueInput := textinput.New()
	valueInput.Placeholder = "Value.."
	valueInput.CharLimit = 50
	valueInput.Width = 50
	valueInput.Prompt = ""

	baseButton := lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0"))
	confirmButton := baseButton.Copy().Background(green)
	cancelButton := baseButton.Copy().Background(muted)
	deleteButton := baseButton.Copy().Background(red)

	return &DetailPage{
		detailModel:       detailDataModel,
		helpMenu:          helpMenu,
		keys:              keys,
		actionsStyle:      actionsStyle,
		detailsStyle:      detailStyle,
		actions:           actions,
		titleStyle:        titleStyle,
		headingStyle:      headingStyle,
		issuesStyle:       issuesStyle,
		keyInput:          keyInput,
		valueInput:        valueInput,
		highlightedButton: confirmButton,
		mutedButton:       cancelButton,
		deleteButton:      deleteButton,
	}
}

func (d *DetailPage) Init() tea.Cmd {
	return nil
}

func (d *DetailPage) getDetails() tea.Msg {
	res, err := d.detailModel.GetAll(d.currentProfile.ID)
	if err != nil {
		return retrieveDetailsMsg{err: err}
	}
	return retrieveDetailsMsg{details: res}
}

func (d *DetailPage) setAliasList() {
}

func (d *DetailPage) setEnvList() {
}

func (d *DetailPage) setActionsList() {
}

func (d *DetailPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DetailStartMsg:
		d.currentProfile = msg.currentProfile
		d.currentUserFlow = retrieveDetails
		return d, func() tea.Msg {
			return d.getDetails()
		}
	case retrieveDetailsMsg:
		if msg.err != nil {
			return d, func() tea.Msg {
				return IssueMsg{Inner: msg.err}
			}
		}
		d.currentUserFlow = listDetails
		d.details = msg.details
		d.activePane = aliasPane
		d.setAliasList()
		d.setEnvList()
		d.setActionsList()
	}
	return d, nil
}

func (d *DetailPage) View() string {
	return "detail page.."
}

func (d *DetailPage) UpdateSize(width, height int) {
	d.width = width
	d.height = height
}
