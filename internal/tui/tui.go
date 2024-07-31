package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Run(debugFlag bool) error {
	model := NewMaggiModel(debugFlag)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		return err
	}
	return nil
}
