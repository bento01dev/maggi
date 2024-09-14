package tui

import (
	"github.com/bento01dev/maggi/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(debugFlag bool, maggiRespository *data.MaggiRepository) error {
	model := NewMaggiModel(debugFlag, maggiRespository)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		return err
	}
	return nil
}
