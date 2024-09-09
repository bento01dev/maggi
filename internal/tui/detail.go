package tui

import (
	"fmt"
	"strings"

	"github.com/bento01dev/maggi/internal/data"
	"github.com/charmbracelet/bubbles/cursor"
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

type DetailDoneMsg struct {}

func (d DetailDoneMsg) Next() pageType {
    return profile
}

type retrieveDetailsMsg struct {
	details []data.Detail
	err     error
}

type detailEditedMsg struct{}

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
	editDetailKey
	editDetailValue
	editDetailConfirm
	editDetailCancel
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
	Add(key string, value string, detailType data.DetailType, profileID int) (*data.Detail, error)
	Update(detail data.Detail, key string, value string) (*data.Detail, error)
	Delete(detail data.Detail) error
}

type DetailPage struct {
	currentDetail     *data.Detail
	emptyDisplay      bool
	infoFlag          bool
	isErrInfo         bool
	width             int
	height            int
	currentUserFlow   detailsUserFlow
	detailType        detailType
	currentStage      detailStage
	activePane        detailPagePane
	infoMsg           string
	currentProfile    data.Profile
	detailModel       detailModel
	details           []data.Detail
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
	keyInputStyle     lipgloss.Style
	valueInputStyle   lipgloss.Style
	keyTextArea       textarea.Model
	valueTextArea     textarea.Model
	aliasList         list.Model
	envList           list.Model
	actionsList       list.Model
	keyInput          textinput.Model
	valueInput        textinput.Model
	highlightedButton lipgloss.Style
	mutedButton       lipgloss.Style
	redButton         lipgloss.Style
	confirmButton     lipgloss.Style
	cancelButton      lipgloss.Style
	deleteButton      lipgloss.Style
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
	keyInputStyle := lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(3).PaddingRight(1)
	valueInputStyle := lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1)

	baseButton := lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0"))
	confirmButton := baseButton.Copy().Background(green)
	mutedButton := baseButton.Copy().Background(muted)
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
		keyInputStyle:     keyInputStyle,
		valueInputStyle:   valueInputStyle,
		keyTextArea:       createTextArea(true),
		valueTextArea:     createTextArea(true),
		keyInput:          createTextInput(true, "Enter Name..", "", 50, keyInputStyle),
		valueInput:        createTextInput(true, "Enter Value..", "", 50, valueInputStyle),
		highlightedButton: confirmButton,
		mutedButton:       mutedButton,
		redButton:         deleteButton,
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
	if enabled {
		input.PromptStyle = baseStyle.Copy().Foreground(green)
		input.TextStyle = baseStyle.Copy().Foreground(green)
		return input
	}
	input.PromptStyle = baseStyle.Copy().Foreground(muted)
	input.TextStyle = baseStyle.Copy().Foreground(muted)
	return input
}

func (d *DetailPage) getDetails() tea.Msg {
	res, err := d.detailModel.GetAll(d.currentProfile.ID)
	if err != nil {
		return retrieveDetailsMsg{err: err}
	}
	return retrieveDetailsMsg{details: res}
}

func (d *DetailPage) addDetail(key, value string) (*data.Detail, error) {
	var dataDetailType data.DetailType
	switch d.detailType {
	case detailTypeAlias:
		dataDetailType = data.AliasDetail
	case detailTypeEnv:
		dataDetailType = data.EnvDetail
	}
	detail, err := d.detailModel.Add(key, value, dataDetailType, d.currentProfile.ID)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (d *DetailPage) updateDetail(key, value string) (*data.Detail, error) {
	detail, err := d.detailModel.Update(*d.currentDetail, key, value)
	if err != nil {
		return nil, err
	}
	return detail, err
}

func (d *DetailPage) deleteDetail() error {
	err := d.detailModel.Delete(*d.currentDetail)
	return err
}

func (d *DetailPage) resetDetails() error {
	details, err := d.detailModel.GetAll(d.currentProfile.ID)
	if err != nil {
		return err
	}
	d.details = details
	if len(d.details) == 0 {
		d.emptyDisplay = true
	}
	return err
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
	confirmButton := d.mutedButton
	deleteButton := d.mutedButton
	cancelButton := d.mutedButton

	// first styles are updated at pane level. then styles are updated depending on stage.
	// the order of the switch statements is important for that reason

	switch d.activePane {
	case envPane:
		envStyle = envStyle.Copy().BorderForeground(green)
	case aliasPane:
		aliasStyle = aliasStyle.Copy().BorderForeground(green)
	case detailDisplayPane:
		displayStyle = displayStyle.Copy().BorderForeground(green)
		enabled = true
	case detailActionPane:
		enabled = true
		actionsStyle = actionsStyle.Copy().BorderForeground(green)
	}

	switch d.currentStage {
	case editDetailKey:
		keyDisplayStyle = keyDisplayStyle.Copy().Foreground(blue)
	case editDetailValue:
		valueDisplayStyle = valueDisplayStyle.Copy().Foreground(blue)
	case editDetailConfirm:
		confirmButton = d.highlightedButton
	case editDetailCancel:
		cancelButton = d.highlightedButton
	case deleteDetailConfirm:
		deleteButton = d.redButton
	case deleteDetailCancel:
		cancelButton = d.highlightedButton
	}

	d.displayStyle = displayStyle
	d.actionsStyle = actionsStyle
	d.aliasStyle = aliasStyle
	d.envStyle = envStyle
	d.keyDisplayStyle = keyDisplayStyle
	d.valueDisplayStyle = valueDisplayStyle
	d.keyTextArea = createTextArea(enabled)
	d.valueTextArea = createTextArea(enabled)
	d.keyInput = createTextInput(enabled, d.keyInput.Placeholder, d.keyInput.Value(), 50, d.keyInputStyle)
	d.valueInput = createTextInput(enabled, d.valueInput.Placeholder, d.valueInput.Value(), 50, d.valueInputStyle)
	d.confirmButton = confirmButton
	d.cancelButton = cancelButton
	d.deleteButton = deleteButton
}

func (d *DetailPage) handleDisplayPaneEvent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch d.currentStage {
	case editDetailKey:
		d.keyInput, cmd = d.keyInput.Update(msg)
	case editDetailValue:
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
			d.emptyDisplay = true
		} else {
			d.emptyDisplay = false
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
			d.emptyDisplay = true
		} else {
			d.emptyDisplay = false
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

func (d *DetailPage) resetInfoBag() {
	d.infoFlag = false
	d.isErrInfo = false
	d.infoMsg = ""
}

func (d *DetailPage) checkIfKeyExists(key string) bool {
	var exists bool
	if d.currentUserFlow == newDetail {
		for _, detail := range d.details {
			if detail.Key == key {
				return true
			}
		}
	}
	if d.currentUserFlow == updateDetail {
		for _, detail := range d.details {
			if detail.Key == key && detail.ID != d.currentDetail.ID {
				return true
			}
		}
	}
	return exists
}

func (d *DetailPage) handleEsc() tea.Cmd {
    if d.currentUserFlow == listDetails {
        return func() tea.Msg {
            return DetailDoneMsg{}
        }
    }
	d.currentUserFlow = listDetails
	d.activePane = envPane
	d.currentDetail = nil
    d.emptyDisplay = true
	d.infoMsg = ""
	d.isErrInfo = false
	d.infoFlag = false
	d.keyInput.SetValue("")
	d.valueInput.SetValue("")
	d.keyTextArea.SetValue("")
	d.valueTextArea.SetValue("")
	d.updatePaneStyles()
    return nil
}

func (d *DetailPage) setTextAreaValues() {
	if d.currentDetail == nil {
		return
	}
	d.keyTextArea.SetValue(d.currentDetail.Key)
	d.valueTextArea.SetValue(d.currentDetail.Value)
}

func (d *DetailPage) getCmdForStage() tea.Cmd {
	switch d.currentStage {
	case editDetailKey:
		return tea.Batch(d.keyInput.Focus(), d.keyInput.Cursor.BlinkCmd())
	case editDetailValue:
		return tea.Batch(d.valueInput.Focus(), d.valueInput.Cursor.BlinkCmd())
	default:
		return nil
	}
}

func (d *DetailPage) handleTab(shift bool) tea.Cmd {
	switch d.currentUserFlow {
	case listDetails, viewDetail:
		d.handleListDetailsTab(shift)
	case newDetail, updateDetail:
		d.handleEditDetailTab(shift)
	case deleteDetail:
		d.handleDeleteDetailTab(shift)
	}
	d.setActionsList()
	d.updatePaneStyles()
	d.setTextAreaValues()
	return d.getCmdForStage()
}

func (d *DetailPage) handleListDetailsTab(shift bool) {
	if shift {
		switch d.activePane {
		case envPane:
			d.activePane = aliasPane
			d.detailType = detailTypeAlias
		case aliasPane:
			if d.emptyDisplay {
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
		}
		return
	}

	switch d.activePane {
	case envPane:
		if d.emptyDisplay {
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
	case detailDisplayPane:
		d.activePane = detailActionPane
	}
}

func (d *DetailPage) handleEditDetailTab(shift bool) {
	if shift {
		switch d.activePane {
		case envPane:
			d.activePane = aliasPane
		case aliasPane:
			d.activePane = detailActionPane
			d.currentStage = editDetailCancel
		case detailActionPane:
			switch d.currentStage {
			case editDetailConfirm:
				d.activePane = detailDisplayPane
				d.currentStage = editDetailValue
			case editDetailCancel:
				d.currentStage = editDetailConfirm
			}
		case detailDisplayPane:
			switch d.currentStage {
			case editDetailKey:
				d.activePane = envPane
				d.currentStage = chooseDetailAction
			case editDetailValue:
				d.currentStage = editDetailKey
			}
		}
		return
	}

	switch d.activePane {
	case envPane:
		d.activePane = detailDisplayPane
		d.currentStage = editDetailKey
	case aliasPane:
		d.activePane = envPane
	case detailDisplayPane:
		switch d.currentStage {
		case editDetailKey:
			d.currentStage = editDetailValue
		case editDetailValue:
			d.currentStage = editDetailConfirm
			d.activePane = detailActionPane
		}
	case detailActionPane:
		switch d.currentStage {
		case editDetailConfirm:
			d.currentStage = editDetailCancel
		case editDetailCancel:
			d.activePane = aliasPane
			d.currentStage = chooseDetailAction
		}
	}
}

func (d *DetailPage) handleDeleteDetailTab(shift bool) {
	if shift {
		switch d.activePane {
		case envPane:
			d.activePane = aliasPane
		case aliasPane:
			d.activePane = detailActionPane
			d.currentStage = deleteDetailConfirm
		case detailActionPane:
			switch d.currentStage {
			case deleteDetailConfirm:
				d.currentStage = deleteDetailCancel
			case deleteDetailCancel:
				d.activePane = detailDisplayPane
				d.currentStage = deleteDetailView
			}
		case detailDisplayPane:
			d.activePane = envPane
			d.currentStage = chooseDetailAction
		}
		return
	}

	switch d.activePane {
	case envPane:
		d.activePane = detailDisplayPane
		d.currentStage = deleteDetailView
	case aliasPane:
		d.activePane = envPane
	case detailDisplayPane:
		d.activePane = detailActionPane
		d.currentStage = deleteDetailConfirm
	case detailActionPane:
		switch d.currentStage {
		case deleteDetailConfirm:
			d.currentStage = deleteDetailCancel
		case deleteDetailCancel:
			d.activePane = aliasPane
			d.currentStage = chooseDetailAction
		}
	}
}

func (d *DetailPage) handleEnter() tea.Cmd {
	var cmd tea.Cmd
	switch d.currentUserFlow {
	case listDetails, viewDetail:
		cmd = d.handleListDetailsEnter()
	case newDetail, updateDetail:
		cmd = d.handleEditDetailEnter()
	case deleteDetail:
		cmd = d.handleDeleteDetailEnter()
	default:
		return nil
	}

	return cmd
}

func (d *DetailPage) handleCancel() {
	d.currentStage = chooseDetailAction
	switch d.detailType {
	case detailTypeAlias:
		d.activePane = aliasPane
	case detailTypeEnv:
		d.activePane = envPane
	default:
		d.activePane = envPane
	}
	d.emptyDisplay = true
	d.keyInput.SetValue("")
	d.valueInput.SetValue("")
	d.keyTextArea.SetValue("")
	d.valueTextArea.SetValue("")
	d.currentUserFlow = listDetails
	d.updatePaneStyles()
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
			d.currentStage = editDetailKey
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
		d.currentStage = editDetailKey
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
		d.currentStage = editDetailKey
	}
	d.setActionsList()
	d.updatePaneStyles()
	d.setTextAreaValues()
	return tea.Batch(d.keyInput.Focus(), d.keyInput.Cursor.BlinkCmd())
}

func (d *DetailPage) handleEditDetailEnter() tea.Cmd {
	switch d.activePane {
	case aliasPane, envPane:
		return d.handleListDetailsEnter()
	}

	d.resetInfoBag()
	switch d.currentStage {
	case editDetailKey:
		key := strings.TrimSpace(d.keyInput.Value())
		if key == "" {
			d.infoFlag = true
			d.isErrInfo = true
			d.infoMsg = "Please pass a valid key. You can exit flow by pressing <esc> if needed"
			return tea.Batch(d.keyInput.Focus(), d.keyInput.Cursor.BlinkCmd())
		}

		if d.checkIfKeyExists(key) {
			d.infoFlag = true
			d.isErrInfo = true
			d.infoMsg = fmt.Sprintf("Key %s already exists in profile. You can <esc> to edit or delete the existing entry before creating a new one!", key)
			return tea.Batch(d.keyInput.Focus(), d.keyInput.Cursor.BlinkCmd())
		}
		d.currentStage = editDetailValue
		d.updatePaneStyles()
		return tea.Batch(d.valueInput.Focus(), d.valueInput.Cursor.BlinkCmd())
	case editDetailValue:
		key := strings.TrimSpace(d.keyInput.Value())
		if key == "" {
			d.infoFlag = true
			d.isErrInfo = true
			d.infoMsg = "Please pass a valid key. You can exit flow by pressing <esc> if needed"
			d.currentStage = editDetailKey
			d.valueInput.Cursor.SetMode(cursor.CursorHide)
			return tea.Batch(d.keyInput.Focus(), d.keyInput.Cursor.BlinkCmd())
		}

		if d.checkIfKeyExists(key) {
			d.infoFlag = true
			d.isErrInfo = true
			d.infoMsg = fmt.Sprintf("Key %s already exists in profile. You can <esc> to edit or delete the existing entry before creating a new one!", key)
			d.currentStage = editDetailKey
			d.valueInput.Cursor.SetMode(cursor.CursorHide)
			return tea.Batch(d.keyInput.Focus(), d.keyInput.Cursor.BlinkCmd())
		}

		value := strings.TrimSpace(d.valueInput.Value())
		if value == "" {
			d.infoFlag = true
			d.isErrInfo = true
			d.infoMsg = "Please pass a valid value. You can exit flow by pressing <esc> if needed"
			return tea.Batch(d.valueInput.Focus(), d.valueInput.Cursor.BlinkCmd())
		}
		d.currentStage = editDetailConfirm
		d.activePane = detailActionPane
		d.updatePaneStyles()
		return nil
	case editDetailConfirm:
		key := strings.TrimSpace(d.keyInput.Value())
		value := strings.TrimSpace(d.valueInput.Value())
		if key == "" {
			d.infoFlag = true
			d.isErrInfo = true
			d.infoMsg = "Please pass a valid key. You can exit flow by pressing <esc> if needed"
			d.currentStage = editDetailKey
			d.valueInput.Cursor.SetMode(cursor.CursorHide)
			return tea.Batch(d.keyInput.Focus(), d.keyInput.Cursor.BlinkCmd())
		}

		if d.checkIfKeyExists(key) {
			d.infoFlag = true
			d.isErrInfo = true
			d.infoMsg = fmt.Sprintf("Key %s already exists in profile. You can <esc> to edit or delete the existing entry before creating a new one!", key)
			d.currentStage = editDetailKey
			d.valueInput.Cursor.SetMode(cursor.CursorHide)
			return tea.Batch(d.keyInput.Focus(), d.keyInput.Cursor.BlinkCmd())
		}

		if value == "" {
			d.infoFlag = true
			d.isErrInfo = true
			d.infoMsg = "Please pass a valid value. You can exit flow by pressing <esc> if needed"
			return tea.Batch(d.valueInput.Focus(), d.valueInput.Cursor.BlinkCmd())
		}

		return func() tea.Msg {
			var err error
			switch d.currentUserFlow {
			case newDetail:
				_, err = d.addDetail(key, value)
			case updateDetail:
				_, err = d.updateDetail(key, value)
			}
			if err != nil {
				return IssueMsg{Inner: err}
			}
			return detailEditedMsg{}
		}
	case editDetailCancel:
		d.handleCancel()
		return nil
	}
	return nil
}

func (d *DetailPage) handleDeleteDetailEnter() tea.Cmd {
	switch d.activePane {
	case aliasPane, envPane:
		return d.handleListDetailsEnter()
	}

	switch d.currentStage {
	case deleteDetailView:
		d.currentStage = deleteDetailConfirm
		d.activePane = detailActionPane
		d.updatePaneStyles()
		d.setTextAreaValues()
		return nil
	case deleteDetailCancel:
		d.handleCancel()
		return nil
	case deleteDetailConfirm:
		return func() tea.Msg {
			err := d.deleteDetail()
			if err != nil {
				return IssueMsg{Inner: err}
			}
			return detailEditedMsg{}
		}
	}
	return nil
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
	if d.emptyDisplay {
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

func (d *DetailPage) viewEditDetail() string {
	var confirmButtonStr string
	switch d.currentUserFlow {
	case newDetail:
		confirmButtonStr = "Create"
	case updateDetail:
		confirmButtonStr = "Update"
	default:
		confirmButtonStr = "Create"
	}

	if d.infoFlag {
		infoStyle := d.issuesStyle.Copy().BorderForeground(green)
		if d.isErrInfo {
			infoStyle = d.issuesStyle.Copy().BorderForeground(red)
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
						infoStyle.Render(d.infoMsg),
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
						d.actionsStyle.Render(
							lipgloss.JoinHorizontal(
								lipgloss.Left,
								d.confirmButton.Render(confirmButtonStr),
								d.cancelButton.Render("Cancel"),
							),
						),
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
								d.keyInput.View(),
							),
							lipgloss.JoinHorizontal(
								lipgloss.Left,
								d.valueDisplayStyle.Render("Value: "),
								d.valueInput.View(),
							),
						),
					),
					d.actionsStyle.Render(
						lipgloss.JoinHorizontal(
							lipgloss.Left,
							d.confirmButton.Render(confirmButtonStr),
							d.cancelButton.Render("Cancel"),
						),
					),
				),
			),
			d.helpMenu.View(d.keys),
		),
	)
}

func (d *DetailPage) viewDeleteDetail() string {
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
					d.actionsStyle.Render(
						lipgloss.JoinHorizontal(
							lipgloss.Left,
							d.deleteButton.Render("Delete"),
							d.cancelButton.Render("Cancel"),
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
			return d, d.handleTab(false)
		case tea.KeyShiftTab:
			return d, d.handleTab(true)
		case tea.KeyEnter:
			return d, d.handleEnter()
		case tea.KeyEsc:
			// d.handleEsc()
			return d, d.handleEsc()
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
		d.emptyDisplay = true
		d.setDetailLists()
		d.setActionsList()
	case detailEditedMsg:
		err := d.resetDetails()
		if err != nil {
			return d, func() tea.Msg {
				return IssueMsg{Inner: err}
			}
		}
		switch d.detailType {
		case detailTypeAlias:
			d.activePane = aliasPane
		case detailTypeEnv:
			d.activePane = envPane
		default:
			d.activePane = envPane
			d.detailType = detailTypeEnv
		}
		d.currentUserFlow = listDetails
		d.currentStage = chooseDetailAction
		d.emptyDisplay = true

		d.keyInput.SetValue("")
		d.valueInput.SetValue("")

		d.setDetailLists()
		d.updatePaneStyles()
		return d, nil
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
	case newDetail, updateDetail:
		return d.viewEditDetail()
	case deleteDetail:
		return d.viewDeleteDetail()
	}
	return "detail page.."
}

func (d *DetailPage) UpdateSize(width, height int) {
	d.width = width
	d.height = height
}
