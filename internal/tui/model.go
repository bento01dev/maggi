package tui

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/bento01dev/maggi/internal/data"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	green               = lipgloss.Color("#04b575")
	yellow              = lipgloss.Color("#ffd866")
	red                 = lipgloss.Color("#ff6188")
	blue                = lipgloss.Color("#2ea0f9")
	muted               = lipgloss.Color("241")
	selectedItemStyle   = lipgloss.NewStyle().PaddingLeft(4).Foreground(blue)
	unselectedItemStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(muted)
)

type pageType int

const (
	start pageType = iota
	profile
	detail
	issue
)

type Page interface {
	tea.Model
	UpdateSize(width, height int)
}

type PageTurner interface {
	Next() pageType
}

type GenericTurner pageType

func (g GenericTurner) Next() pageType {
	return pageType(g)
}

type ListRenderFunc func(list.Item) string

type ListItemDelegate struct {
	RenderFunc ListRenderFunc
}

func (i ListItemDelegate) Height() int                             { return 1 }
func (i ListItemDelegate) Spacing() int                            { return 0 }
func (i ListItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (i ListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	itemStr := i.RenderFunc(listItem)
	if itemStr == "" {
		return
	}

	styleFn := unselectedItemStyle.Render
	if index == m.Index() {
		styleFn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, styleFn(i.RenderFunc(listItem)))
}

func GenerateList(items []list.Item, renderFunc ListRenderFunc, width int, height int, filtering bool) list.Model {
	l := list.New(items, ListItemDelegate{RenderFunc: renderFunc}, width, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(filtering)
	l.SetFilteringEnabled(filtering)
	if filtering {
		valueInput := textinput.New()
		valueInput.Placeholder = "Search.."
		valueInput.CharLimit = 50
		valueInput.Width = width
		valueInput.Prompt = "> "
		valueInput.PromptStyle = lipgloss.NewStyle().Foreground(blue)
		valueInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(muted)
		valueInput.TextStyle = lipgloss.NewStyle().Foreground(blue)
		l.FilterInput = valueInput
	}
	l.SetShowPagination(false)
	l.SetShowHelp(false)
	return l
}

type MaggiModel struct {
	currentPage pageType
	pages       map[pageType]Page
	profile     data.Profile
	err         error
}

func NewMaggiModel(debugFlag bool, dataModel data.DataModel) *MaggiModel {
	return &MaggiModel{
		pages: map[pageType]Page{
			issue:   NewIssuePage(debugFlag),
			profile: NewProfilePage(dataModel.Profiles),
			detail:  NewDetailPage(dataModel.Details),
		},
	}
}

func (m *MaggiModel) Init() tea.Cmd {
	m.currentPage = start
	return func() tea.Msg {
		return GenericTurner(profile)
	}
}

func (m *MaggiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for _, page := range m.pages {
			page.UpdateSize(msg.Width, msg.Height)
		}
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case PageTurner:
		m.handlePageTransition(msg)
		return m, m.pageInitCmd()
	}
	page, ok := m.pages[m.currentPage]
	if !ok {
		m.err = errors.New("unknown page lookup")
		m.currentPage = issue
		return m, nil
	}
	_, cmd = page.Update(msg)
	return m, cmd
}

func (m *MaggiModel) handlePageTransition(msg PageTurner) {
	switch msg := msg.(type) {
	case ProfileDoneMsg:
		m.profile = msg.profile
	}
	m.currentPage = msg.Next()
}

func (m *MaggiModel) pageInitCmd() tea.Cmd {
	var msg tea.Msg
	switch m.currentPage {
	case profile:
		msg = ProfileStartMsg{}
	case detail:
		msg = DetailStartMsg{currentProfile: m.profile}
	}
	return func() tea.Msg {
		return msg
	}
}

func (m *MaggiModel) View() string {
	page, ok := m.pages[m.currentPage]
	if !ok {
		return "something horribly wrong with the app"
	}
	return page.View()
}
