package tui

import (
	"fmt"
	"strings"

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
	defaultDPWidth       int = 120
	defaultSideBarWidth  int = 30
	defaultDisplayWidth  int = 90
	defaultDisplayHeight int = 7
	defaultDPHeight      int = 20
	defaultSideBarHeight int = 5
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
	detailDisplayPane
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
	addDetailKey
	addDetailValue
	addDetailConfirm
	addDetailCancel
	updateDetailKey
	updateDetailValue
	updateDetailConfirm
	updateDetailCancel
	deleteDetailView
	deleteDetailConfirm
	deleteDetailCancel
)

type detailActionItem struct {
	next        detailsUserFlow
	description string
}

func (a detailActionItem) FilterValue() string { return "" }
func renderDetailActionItem(i list.Item) string {
	a, ok := i.(detailActionItem)
	if !ok {
		return ""
	}
	return a.description
}

type detailItem struct {
	id     int
	key    string
	value  string
	action bool
}

func (d detailItem) FilterValue() string {
	if d.action {
		return ""
	}
	return d.key
}
func renderDetailItem(i list.Item) string {
	p, ok := i.(detailItem)
	if !ok {
		return ""
	}
	return p.key
}

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
	newDetailOption   bool
	width             int
	height            int
	currentUserFlow   detailsUserFlow
	detailType        detailType
	currentStage      detailStage
	activePane        detailPagePane
	currentProfile    data.Profile
	detailModel       detailModel
	details           []data.Detail
	currentDetail     data.Detail
	helpMenu          help.Model
	keys              detailHelpKeys
	titleStyle        lipgloss.Style
	headingStyle      lipgloss.Style
	issuesStyle       lipgloss.Style
	actionsStyle      lipgloss.Style
	aliasStyle        lipgloss.Style
	envStyle          lipgloss.Style
	displayStyle      lipgloss.Style
	aliasList         list.Model
	envList           list.Model
	actionsList       list.Model
	keyInput          textinput.Model
	valueInput        textinput.Model
	highlightedButton lipgloss.Style
	mutedButton       lipgloss.Style
	deleteButton      lipgloss.Style
	infoMsg           string
	actions           []string
}

// TODO: the styling elements and help menu are duplicates of what is in profile. de-duplicate if needed later.
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

	actionsStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding()
	displayStyle := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding()
	aliasStyle := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding()
	envStyle := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding()
	issuesStyle := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).UnsetPadding().BorderForeground(red)
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
		titleStyle:        titleStyle,
		headingStyle:      headingStyle,
		issuesStyle:       issuesStyle,
		displayStyle:      displayStyle,
		aliasStyle:        aliasStyle,
		envStyle:          envStyle,
		keyInput:          keyInput,
		valueInput:        valueInput,
		highlightedButton: confirmButton,
		mutedButton:       cancelButton,
		deleteButton:      deleteButton,
	}
}

func (d *DetailPage) getDetails() tea.Msg {
	// res, err := d.detailModel.GetAll(d.currentProfile.ID)
	// if err != nil {
	// 	return retrieveDetailsMsg{err: err}
	// }
	// return retrieveDetailsMsg{details: res}
	res := []data.Detail{
		{
			ID:         1,
			Key:        "ENV_1",
			Value:      "VALUE_1",
			DetailType: data.EnvDetail,
			ProfileID:  1,
		},
		{
			ID:         2,
			Key:        "ENV_2",
			Value:      "VALUE_2",
			DetailType: data.EnvDetail,
			ProfileID:  1,
		},
		{
			ID:         3,
			Key:        "ALIAS_1",
			Value:      "VALUE_1",
			DetailType: data.AliasDetail,
			ProfileID:  1,
		},
		{
			ID:         4,
			Key:        "ALIAS_2",
			Value:      "VALUE_2",
			DetailType: data.AliasDetail,
			ProfileID:  1,
		},
	}
	return retrieveDetailsMsg{details: res}
}

func (d *DetailPage) setDetailLists() {
	aliasList := []list.Item{}
	envList := []list.Item{}
	aliasList = append(aliasList, detailItem{key: "Add alias...", action: true})
	envList = append(envList, detailItem{key: "Add env var...", action: true})
	for _, detail := range d.details {
		switch detail.DetailType {
		case data.AliasDetail:
			aliasList = append(aliasList, detailItem{id: detail.ID, key: detail.Key, value: detail.Value})
		case data.EnvDetail:
			envList = append(envList, detailItem{id: detail.ID, key: detail.Key, value: detail.Value})
		}
	}
	d.aliasList = GenerateList(aliasList, renderDetailItem, defaultSideBarWidth, defaultSideBarHeight, true)
	d.envList = GenerateList(envList, renderDetailItem, defaultSideBarWidth, defaultSideBarHeight, true)
	d.updatePaneStyles()
}

func (d *DetailPage) setActionsList() {
	var actionItems []list.Item
	switch d.detailType {
	case detailTypeAlias:
		actionItems = []list.Item{
			detailActionItem{
				description: "Update Alias",
				next:        updateDetail,
			},
			detailActionItem{
				description: "Delete Alias",
				next:        deleteDetail,
			},
		}
	case detailTypeEnv:
		actionItems = []list.Item{
			detailActionItem{
				description: "Update Env",
				next:        updateDetail,
			},
			detailActionItem{
				description: "Delete Env",
				next:        deleteDetail,
			},
		}
	}
	d.actionsList = GenerateList(actionItems, renderDetailActionItem, defaultDisplayWidth, 2, false)
	d.updatePaneStyles()
}

func (d *DetailPage) updatePaneStyles() {
	displayStyle := d.displayStyle.Copy().BorderForeground(muted)
	actionsStyle := d.actionsStyle.Copy().BorderForeground(muted)
	aliasStyle := d.aliasStyle.Copy().BorderForeground(muted)
	envStyle := d.envStyle.Copy().BorderForeground(muted)

	switch d.activePane {
	case envPane:
		envStyle = envStyle.Copy().BorderForeground(green)
	case aliasPane:
		aliasStyle = aliasStyle.Copy().BorderForeground(green)
	case detailDisplayPane:
		displayStyle = displayStyle.Copy().BorderForeground(green)
	case detailActionPane:
		actionsStyle = actionsStyle.Copy().BorderForeground(green)
	}

	d.displayStyle = displayStyle
	d.actionsStyle = actionsStyle
	d.aliasStyle = aliasStyle
	d.envStyle = envStyle
}

func (d *DetailPage) handleEvent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch d.activePane {
	case envPane:
		d.envList, cmd = d.envList.Update(msg)
		item, ok := d.envList.SelectedItem().(detailItem)
		if !ok {
			return tea.Quit
		}
		if item.action {
			d.newDetailOption = true
		} else {
			d.newDetailOption = false
		}
	case aliasPane:
		d.aliasList, cmd = d.aliasList.Update(msg)
		item, ok := d.aliasList.SelectedItem().(detailItem)
		if !ok {
			return tea.Quit
		}
		if item.action {
			d.newDetailOption = true
		} else {
			d.newDetailOption = false
		}
	}
	return cmd
}

func (d *DetailPage) generateTitle() string {
	first := strings.Repeat("-", 3)
	var second string
	switch d.currentUserFlow {
	case listDetails:
		second = " Details "
	case newDetail:
		switch d.detailType {
		case detailTypeEnv:
			second = " New Env "
		case detailTypeAlias:
			second = " New Alias "
		}
	case updateDetail:
		switch d.detailType {
		case detailTypeEnv:
			second = fmt.Sprintf(" %s | Update Env | %s ", d.currentProfile.Name, d.currentDetail.Key)
		case detailTypeAlias:
			second = fmt.Sprintf(" %s | Update Alias | %s ", d.currentProfile.Name, d.currentDetail.Key)
		}
	case deleteDetail:
		switch d.detailType {
		case detailTypeEnv:
			second = fmt.Sprintf(" %s | Delete Env | %s ", d.currentProfile.Name, d.currentDetail.Key)
		case detailTypeAlias:
			second = fmt.Sprintf(" %s | Delete Alias | %s ", d.currentProfile.Name, d.currentDetail.Key)
		}
	}
	third := strings.Repeat("-", (defaultDPWidth - (len(second) + 3)))
	return first + second + third
}

func (d *DetailPage) generateHeading(name string) string {
	first := strings.Repeat("-", 3)
	second := fmt.Sprintf(" %s ", name)
	third := strings.Repeat("-", (defaultSideBarWidth - (len(second) + 3)))
	return first + second + third
}

func (d *DetailPage) viewListDetails() string {
	if d.newDetailOption {
		return lipgloss.Place(
			d.width,
			d.height,
			lipgloss.Center,
			lipgloss.Center,
			lipgloss.JoinVertical(
				lipgloss.Center,
				d.titleStyle.Render(d.generateTitle()),
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					lipgloss.JoinVertical(
						lipgloss.Center,
						d.headingStyle.Render(d.generateHeading("Envs")),
						d.envStyle.Render(d.envList.View()),
						d.headingStyle.Render(d.generateHeading("Aliases")),
						d.aliasStyle.Render(d.aliasList.View()),
					),
					lipgloss.JoinVertical(
						lipgloss.Center,
						"",
						d.displayStyle.Render(""),
					),
				),
				d.helpMenu.View(d.keys),
			),
		)
	}

	return lipgloss.Place(
		d.width,
		d.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			d.titleStyle.Render(d.generateTitle()),
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.JoinVertical(
					lipgloss.Center,
					d.headingStyle.Render(d.generateHeading("Envs")),
					d.envStyle.Render(d.envList.View()),
					d.headingStyle.Render(d.generateHeading("Aliases")),
					d.aliasStyle.Render(d.aliasList.View()),
				),
				lipgloss.JoinVertical(
					lipgloss.Center,
					"",
					d.displayStyle.Render(""),
					d.actionsStyle.Render(d.actionsList.View()),
				),
			),
			d.helpMenu.View(d.keys),
		),
	)
}

func (d *DetailPage) Init() tea.Cmd {
	return nil
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
		d.activePane = envPane
		d.detailType = detailTypeEnv
		d.newDetailOption = true
		d.setDetailLists()
		d.setActionsList()
	}
	cmd := d.handleEvent(msg)
	return d, cmd
}

func (d *DetailPage) View() string {
	switch d.currentUserFlow {
	case listDetails:
		return d.viewListDetails()
	}
	return "detail page.."
}

func (d *DetailPage) UpdateSize(width, height int) {
	d.width = width
	d.height = height
}
