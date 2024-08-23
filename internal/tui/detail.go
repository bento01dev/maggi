package tui

import (
	"github.com/bento01dev/maggi/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

type DetailStartMsg struct {
	currentProfile data.Profile
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

type detailModel interface {
    GetAll() ([]data.Detail, error)
    Add(key string, value string, detailType data.DetailType, profileID int) (data.Detail, error)
    Update(detail data.Detail, key string, value string) (data.Detail, error)
    Delete(detail data.Detail) error
    DeleteAll(profileID int) error
}

type DetailPage struct {
	width          int
	height         int
	currentProfile data.Profile
	detailModel    detailModel
}

func NewDetailPage(detailDataModel detailModel) *DetailPage {
	return &DetailPage{detailModel: detailDataModel}
}

func (d *DetailPage) Init() tea.Cmd {
	return nil
}

func (d *DetailPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DetailStartMsg:
		d.currentProfile = msg.currentProfile
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
