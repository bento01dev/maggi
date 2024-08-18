package tui

import (
	"github.com/bento01dev/maggi/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(debugFlag bool) error {
    db, err := data.Setup()
    if err != nil {
        return err
    }
    
    dataModel := data.NewDataModel(db)
	model := NewMaggiModel(debugFlag, dataModel)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		return err
	}
	return nil
}
