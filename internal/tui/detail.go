package tui

import (
	"fmt"
	"strings"

	"github.com/bento01dev/maggi/internal/data"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
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
	Search     key.Binding
}

func (h detailHelpKeys) ShortHelp() []key.Binding {
	return []key.Binding{h.ToggleView, h.Search, h.Up, h.Down, h.Esc, h.Quit}
}

func (h detailHelpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{h.ToggleView, h.Search, h.Up, h.Down},
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
	currentDetail     *data.Detail
	helpMenu          help.Model
	keys              detailHelpKeys
	titleStyle        lipgloss.Style
	headingStyle      lipgloss.Style
	issuesStyle       lipgloss.Style
	actionsStyle      lipgloss.Style
	aliasStyle        lipgloss.Style
	envStyle          lipgloss.Style
	displayStyle      lipgloss.Style
	keyDisplayStyle   lipgloss.Style
	valueDisplayStyle lipgloss.Style
	keyTextArea       textarea.Model
	valueTextArea     textarea.Model
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
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
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
	keyDisplayStyle := lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1)
	valueDisplayStyle := lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1)

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
		keyDisplayStyle:   keyDisplayStyle,
		valueDisplayStyle: valueDisplayStyle,
		keyTextArea:       createTextArea(true),
		valueTextArea:     createTextArea(true),
		keyInput:          createTextInput(true, "Enter Name..", "", 50, keyDisplayStyle.Copy().Foreground(green)),
		valueInput:        createTextInput(true, "Enter Value..", "", 50, valueDisplayStyle.Copy().Foreground(green)),
		highlightedButton: confirmButton,
		mutedButton:       cancelButton,
		deleteButton:      deleteButton,
	}
}

func createTextArea(enabled bool) textarea.Model {
	ta := textarea.New()
	ta.SetValue("")
	ta.SetWidth(50)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.Blur()
	ta.BlurredStyle.Base.Italic(true)
	if enabled {
		ta.BlurredStyle = textarea.Style{Base: lipgloss.NewStyle().Foreground(blue).BorderStyle(lipgloss.RoundedBorder()).BorderForeground(green).Height(1).Width(50)}
		return ta
	}
	ta.BlurredStyle = textarea.Style{Base: lipgloss.NewStyle().Foreground(muted).BorderStyle(lipgloss.RoundedBorder()).BorderForeground(muted).Height(1).Width(50)}
	return ta
}

func createTextInput(enabled bool, placeholder string, value string, width int, baseStyle lipgloss.Style) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = width
	input.Width = width
	input.Prompt = ""
	if value != "" {
		input.SetValue(value)
	}
    // input.PromptStyle = baseStyle
    input.TextStyle = baseStyle
	return input
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
	keyDisplayStyle := d.keyDisplayStyle.Copy().Foreground(muted)
	valueDisplayStyle := d.valueDisplayStyle.Copy().Foreground(muted)
	var enabled bool

	switch d.activePane {
	case envPane:
		envStyle = envStyle.Copy().BorderForeground(green)
	case aliasPane:
		aliasStyle = aliasStyle.Copy().BorderForeground(green)
	case detailDisplayPane:
		displayStyle = displayStyle.Copy().BorderForeground(green)
		keyDisplayStyle = keyDisplayStyle.Copy().Foreground(blue)
		valueDisplayStyle = valueDisplayStyle.Copy().Foreground(blue)
		enabled = true
	case detailActionPane:
		keyDisplayStyle = keyDisplayStyle.Copy().Foreground(blue)
		valueDisplayStyle = valueDisplayStyle.Copy().Foreground(blue)
		enabled = true
		actionsStyle = actionsStyle.Copy().BorderForeground(green)
	}

	d.displayStyle = displayStyle
	d.actionsStyle = actionsStyle
	d.aliasStyle = aliasStyle
	d.envStyle = envStyle
	d.keyDisplayStyle = keyDisplayStyle
	d.valueDisplayStyle = valueDisplayStyle
	d.keyTextArea = createTextArea(enabled)
	d.valueTextArea = createTextArea(enabled)
	d.keyInput = createTextInput(enabled, d.keyInput.Placeholder, d.keyInput.Value(), 50, keyDisplayStyle.Copy().Foreground(green))
	d.valueInput = createTextInput(enabled, d.valueInput.Placeholder, d.valueInput.Value(), 50, valueDisplayStyle.Copy().Foreground(green))
}

func (d *DetailPage) handleDisplayPaneEvent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch d.currentStage {
	case addDetailKey, updateDetailKey:
		d.keyInput, cmd = d.keyInput.Update(msg)
	case addDetailValue, updateDetailValue:
		d.valueInput, cmd = d.valueInput.Update(msg)
	}
	return cmd
}

func (d *DetailPage) handleEvent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch d.activePane {
	case detailDisplayPane:
		return d.handleDisplayPaneEvent(msg)
	case envPane:
		d.envList, cmd = d.envList.Update(msg)
		item, ok := d.envList.SelectedItem().(detailItem)
		if !ok {
			return nil
		}
		if item.action {
			d.newDetailOption = true
		} else {
			d.newDetailOption = false
			d.setCurrentDetail(item, data.EnvDetail)
			d.setTextAreaValues()
		}
	case aliasPane:
		d.aliasList, cmd = d.aliasList.Update(msg)
		item, ok := d.aliasList.SelectedItem().(detailItem)
		if !ok {
			return nil
		}
		if item.action {
			d.newDetailOption = true
		} else {
			d.newDetailOption = false
			d.setCurrentDetail(item, data.AliasDetail)
			d.setTextAreaValues()
		}
	case detailActionPane:
		d.actionsList, cmd = d.actionsList.Update(msg)
	}
	return cmd
}

func (d *DetailPage) setCurrentDetail(item detailItem, detailType data.DetailType) {
	d.currentDetail = &data.Detail{
		ID:         item.id,
		Key:        item.key,
		Value:      item.value,
		DetailType: detailType,
		ProfileID:  d.currentProfile.ID,
	}
}

func (d *DetailPage) setTextAreaValues() {
	if d.currentDetail == nil {
		return
	}
	d.keyTextArea.SetValue(d.currentDetail.Key)
	d.valueTextArea.SetValue(d.currentDetail.Value)
}

func (d *DetailPage) handleTab(shift bool) {
	switch d.currentUserFlow {
	case listDetails:
		d.handleListDetailsTab(shift)
	}
	d.setActionsList()
	d.updatePaneStyles()
	d.setTextAreaValues()
}

func (d *DetailPage) handleListDetailsTab(shift bool) {
	if shift {
		switch d.activePane {
		case envPane:
			d.activePane = aliasPane
			d.detailType = detailTypeAlias
		case aliasPane:
			if d.newDetailOption {
				d.activePane = envPane
				d.detailType = detailTypeEnv
			} else {
				d.activePane = detailActionPane
			}
		case detailActionPane:
			d.activePane = detailDisplayPane
		case detailDisplayPane:
			d.activePane = envPane
			d.detailType = detailTypeEnv
			d.currentUserFlow = listDetails
		}
		return
	}

	switch d.activePane {
	case envPane:
		if d.newDetailOption {
			d.activePane = aliasPane
			d.detailType = detailTypeAlias
		} else {
			d.activePane = detailDisplayPane
		}
	case aliasPane:
		d.activePane = envPane
		d.detailType = detailTypeEnv
	case detailActionPane:
		d.activePane = aliasPane
		d.detailType = detailTypeAlias
		d.currentUserFlow = listDetails
	case detailDisplayPane:
		d.activePane = detailActionPane
	}
}

func (d *DetailPage) handleEnter() tea.Cmd {
	var cmd tea.Cmd
	switch d.currentUserFlow {
	case listDetails, viewDetail:
		cmd = d.handleListDetailsEnter()
	default:
		return nil
	}

	return cmd
}

func (d *DetailPage) handleListDetailsEnter() tea.Cmd {
	switch d.activePane {
	case detailDisplayPane:
		d.activePane = detailActionPane
		d.setActionsList()
		d.updatePaneStyles()
		d.setTextAreaValues()
		return nil
	case detailActionPane:
		item, ok := d.actionsList.SelectedItem().(detailActionItem)
		if !ok {
			return tea.Quit
		}
		d.currentUserFlow = item.next
		d.activePane = detailDisplayPane
		switch d.currentUserFlow {
		case deleteDetail:
			d.currentStage = deleteDetailView
			d.setActionsList()
			d.updatePaneStyles()
			d.setTextAreaValues()
			return nil
		case updateDetail:
			d.currentStage = updateDetailKey
			d.keyInput.SetValue(d.currentDetail.Key)
			d.valueInput.SetValue(d.currentDetail.Value)
		}
	case aliasPane:
		d.detailType = detailTypeAlias
		item, ok := d.aliasList.SelectedItem().(detailItem)
		if !ok {
			return tea.Quit
		}
		d.activePane = detailDisplayPane
		if !item.action {
			d.setCurrentDetail(item, data.AliasDetail)
			d.currentUserFlow = viewDetail
			d.setActionsList()
			d.updatePaneStyles()
			d.setTextAreaValues()
			return nil
		}
		d.currentUserFlow = newDetail
		d.currentStage = addDetailKey
	case envPane:
		d.detailType = detailTypeEnv
		item, ok := d.envList.SelectedItem().(detailItem)
		if !ok {
			return tea.Quit
		}
		d.activePane = detailDisplayPane
		if !item.action {
			d.setCurrentDetail(item, data.EnvDetail)
			d.currentUserFlow = viewDetail
			d.setActionsList()
			d.updatePaneStyles()
			d.setTextAreaValues()
			return nil
		}
		d.currentUserFlow = newDetail
		d.currentStage = addDetailKey
	}
	d.setActionsList()
	d.updatePaneStyles()
	d.setTextAreaValues()
	return tea.Batch(d.keyInput.Focus(), d.keyInput.Cursor.BlinkCmd())
}

func (d *DetailPage) generateTitle() string {
	first := strings.Repeat("-", 3)
	var second string
	switch d.currentUserFlow {
	case listDetails:
		second = " Details "
	case viewDetail:
		second = fmt.Sprintf(" %s | View | %s ", d.currentProfile.Name, d.currentDetail.Key)
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
					d.displayStyle.Render(
						lipgloss.JoinVertical(
							lipgloss.Left,
							lipgloss.JoinHorizontal(
								lipgloss.Left,
								d.keyDisplayStyle.Render("Name: "),
								d.keyTextArea.View(),
							),
							lipgloss.JoinHorizontal(
								lipgloss.Left,
								d.valueDisplayStyle.Render("Value: "),
								d.valueTextArea.View(),
							),
						),
					),
					d.actionsStyle.Render(d.actionsList.View()),
				),
			),
			d.helpMenu.View(d.keys),
		),
	)
}

// i know.. naming is dumb. will address later if i feel dumb enough
func (d *DetailPage) viewViewDetail() string {
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
					d.displayStyle.Render(
						lipgloss.JoinVertical(
							lipgloss.Left,
							lipgloss.JoinHorizontal(
								lipgloss.Left,
								d.keyDisplayStyle.Render("Name: "),
								d.keyTextArea.View(),
							),
							lipgloss.JoinHorizontal(
								lipgloss.Left,
								d.valueDisplayStyle.Render("Value: "),
								d.valueTextArea.View(),
							),
						),
					),
					d.actionsStyle.Render(d.actionsList.View()),
				),
			),
			d.helpMenu.View(d.keys),
		),
	)
}

func (d *DetailPage) viewNewDetail() string {
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
					d.displayStyle.Render(
						lipgloss.JoinVertical(
							lipgloss.Left,
							lipgloss.JoinHorizontal(
								lipgloss.Left,
								d.keyDisplayStyle.Render("Name: "),
								d.keyInput.View(),
							),
							lipgloss.JoinHorizontal(
								lipgloss.Left,
								d.valueDisplayStyle.Render("Value: "),
								d.valueInput.View(),
							),
						),
					),
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
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			d.handleTab(false)
			return d, nil
		case tea.KeyShiftTab:
			d.handleTab(true)
			return d, nil
		case tea.KeyEnter:
			return d, d.handleEnter()
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
	case viewDetail:
		return d.viewViewDetail()
	case newDetail:
		return d.viewNewDetail()
	}
	return "detail page.."
}

func (d *DetailPage) UpdateSize(width, height int) {
	d.width = width
	d.height = height
}
