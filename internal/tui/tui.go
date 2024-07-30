package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(debugFlag bool) error {
	fmt.Println("maggi tui..")
	model := NewMaggiModel(debugFlag)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		return err
	}
	return nil
}
