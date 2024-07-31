package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

const defaultWrapWidth int = 100

type IssueMsg struct {
	Inner      error
	Message    string
	StackTrace string
}

type helpKeysMap struct {
	Quit key.Binding
}

func (h helpKeysMap) ShortHelp() []key.Binding {
	return []key.Binding{h.Quit}
}

func (h helpKeysMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{h.Quit},
	}
}

type IssuePage struct {
	debugFlag  bool
	width      int
	height     int
	inner      error
	message    string
	stackTrace string
	errorStyle lipgloss.Style
	helpMenu   help.Model
	keysMap    helpKeysMap
}

func NewIssuePage(debugFlag bool) *IssuePage {
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
	keysMap := helpKeysMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}

	errorStyle := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(red)

	return &IssuePage{
		debugFlag:  debugFlag,
		errorStyle: errorStyle,
		helpMenu:   helpMenu,
		keysMap:    keysMap,
	}
}

func (i *IssuePage) Init() tea.Cmd {
	return nil
}

func (i *IssuePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case IssueMsg:
		wrapWidth := defaultWrapWidth
		if i.width < defaultWrapWidth {
			wrapWidth = i.width
		}
		i.inner = msg.Inner
		i.message = wordwrap.String(msg.Message, wrapWidth)
		i.stackTrace = wordwrap.String(msg.StackTrace, wrapWidth)
	}
	return i, nil
}

func (i *IssuePage) View() string {
	if i.debugFlag {
		return lipgloss.Place(
			i.width,
			i.height,
			lipgloss.Center,
			lipgloss.Center,
			lipgloss.JoinVertical(
				lipgloss.Center,
				i.errorStyle.Render(i.message),
				i.errorStyle.Render(i.stackTrace),
				i.helpMenu.View(i.keysMap),
			),
		)
	}
	return lipgloss.Place(
		i.width,
		i.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			i.errorStyle.Render(i.message),
			i.helpMenu.View(i.keysMap),
		),
	)
}

func (i *IssuePage) UpdateSize(width, height int) {
	i.width = width
	i.height = height
}
