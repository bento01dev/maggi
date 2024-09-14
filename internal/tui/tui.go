package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Run(debugFlag bool, maggiRespository MaggiRepository) error {
	model := NewMaggiModel(debugFlag, maggiRespository)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		return err
	}
	return nil
}
