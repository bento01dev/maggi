package tui

import (
	"errors"
	"testing"

	"github.com/bento01dev/maggi/internal/data"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

type detailModelStub struct {
	getAll func(profileID int) ([]data.Detail, error)
	add    func(key string, value string, detailType data.DetailType, profileID int) (*data.Detail, error)
	update func(detail data.Detail, key string, value string) (*data.Detail, error)
	delete func(detail data.Detail) error
}

func (ds detailModelStub) GetAll(profileID int) ([]data.Detail, error) {
	return ds.getAll(profileID)
}

func (ds detailModelStub) Add(key string, value string, detailType data.DetailType, profileID int) (*data.Detail, error) {
	return ds.add(key, value, detailType, profileID)
}

func (ds detailModelStub) Update(detail data.Detail, key, value string) (*data.Detail, error) {
	return ds.update(detail, key, value)
}

func (ds detailModelStub) Delete(detail data.Detail) error {
	return ds.delete(detail)
}

func TestCreateTextArea(t *testing.T) {
	t.Run("text area is muted when enabled is false", func(t *testing.T) {
		res := createTextArea(false)
		assert.Equal(t, res.BlurredStyle.Base.GetBorderTopForeground(), muted)
		assert.Equal(t, res.Prompt, "")
		assert.Equal(t, res.Value(), "")
		assert.False(t, res.ShowLineNumbers)
	})

	t.Run("text area is green when enabled is true", func(t *testing.T) {
		res := createTextArea(true)
		assert.Equal(t, res.BlurredStyle.Base.GetBorderTopForeground(), green)
		assert.Equal(t, res.Prompt, "")
		assert.Equal(t, res.Value(), "")
		assert.False(t, res.ShowLineNumbers)
	})
}

func TestCreateTextInput(t *testing.T) {
	inputStyle := lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(3).PaddingRight(1)
	t.Run("Style is set to green when enabled", func(t *testing.T) {
		res := createTextInput(true, "placeholder", "value", 50, inputStyle)
		assert.Equal(t, res.PromptStyle.GetForeground(), green)
		assert.Equal(t, res.TextStyle.GetForeground(), green)
		assert.Equal(t, res.Prompt, "")
		assert.Equal(t, res.Placeholder, "placeholder")
		assert.Equal(t, res.Value(), "value")
	})

	t.Run("Style is set to green when not enabled", func(t *testing.T) {
		res := createTextInput(false, "placeholder", "value", 50, inputStyle)
		assert.Equal(t, res.PromptStyle.GetForeground(), muted)
		assert.Equal(t, res.TextStyle.GetForeground(), muted)
		assert.Equal(t, res.Prompt, "")
		assert.Equal(t, res.Placeholder, "placeholder")
		assert.Equal(t, res.Value(), "value")
	})
}

func TestGetDetails(t *testing.T) {
	t.Run("Returns details if no error", func(t *testing.T) {
		details := []data.Detail{
			{
				Key:        "key",
				Value:      "value",
				ID:         1,
				ProfileID:  1,
				DetailType: data.AliasDetail,
			},
		}
		detailPage := NewDetailPage(detailModelStub{getAll: func(profileID int) ([]data.Detail, error) { return details, nil }})
		msg := detailPage.getDetails()
		res, ok := msg.(retrieveDetailsMsg)
		assert.True(t, ok)
		assert.Equal(t, res.details, details)
	})

	t.Run("Returns error in retrieveDetailsMsg on error", func(t *testing.T) {
		detailPage := NewDetailPage(detailModelStub{getAll: func(profileID int) ([]data.Detail, error) { return nil, errors.New("error in retrieval") }})
		msg := detailPage.getDetails()
		res, ok := msg.(retrieveDetailsMsg)
		assert.True(t, ok)
		assert.NotNil(t, res.err)
	})
}

func TestAddDetail(t *testing.T) {
	t.Run("Add alias detail", func(t *testing.T) {
		detail := data.Detail{
			Key:        "key",
			Value:      "value",
			ID:         1,
			ProfileID:  1,
			DetailType: data.AliasDetail,
		}
		detailPage := NewDetailPage(detailModelStub{add: func(key, value string, detailType data.DetailType, profileID int) (*data.Detail, error) {
			return &detail, nil
		}})
		res, err := detailPage.addDetail("key", "value")
		assert.Nil(t, err)
		assert.Equal(t, data.AliasDetail, res.DetailType)
	})

	t.Run("Add env detail", func(t *testing.T) {
		detail := data.Detail{
			Key:        "key",
			Value:      "value",
			ID:         1,
			ProfileID:  1,
			DetailType: data.EnvDetail,
		}
		detailPage := NewDetailPage(detailModelStub{add: func(key, value string, detailType data.DetailType, profileID int) (*data.Detail, error) {
			return &detail, nil
		}})
		res, err := detailPage.addDetail("key", "value")
		assert.Nil(t, err)
		assert.Equal(t, data.EnvDetail, res.DetailType)
	})

	t.Run("Return error if error in add", func(t *testing.T) {
		detailPage := NewDetailPage(detailModelStub{add: func(key, value string, detailType data.DetailType, profileID int) (*data.Detail, error) {
			return nil, errors.New("error in add")
		}})
		res, err := detailPage.addDetail("key", "value")
		assert.Nil(t, res)
		assert.NotNil(t, err)
	})
}

func TestUpdateDetail(t *testing.T) {
	t.Run("returns err if update is error", func(t *testing.T) {
		detailPage := NewDetailPage(detailModelStub{update: func(detail data.Detail, key, value string) (*data.Detail, error) {
			return nil, errors.New("error in update")
		}})
		detailPage.currentDetail = &data.Detail{}
		res, err := detailPage.updateDetail("key", "value")
		assert.Nil(t, res)
		assert.NotNil(t, err)
	})
}

func TestDeleteDetail(t *testing.T) {
	t.Run("returns error if delete returns error", func(t *testing.T) {
		detailPage := NewDetailPage(detailModelStub{delete: func(detail data.Detail) error { return errors.New("error in delete") }})
		detailPage.currentDetail = &data.Detail{}
		err := detailPage.deleteDetail()
		assert.NotNil(t, err)
	})
}

func TestResetDetails(t *testing.T) {
	t.Run("set details based on get all result", func(t *testing.T) {
		details := []data.Detail{
			{
				Key:        "key",
				Value:      "value",
				ProfileID:  1,
				ID:         1,
				DetailType: data.AliasDetail,
			},
		}
		detailPage := NewDetailPage(detailModelStub{getAll: func(profileID int) ([]data.Detail, error) { return details, nil }})
		err := detailPage.resetDetails()
		assert.Nil(t, err)
		assert.Equal(t, details, detailPage.details)
	})

	t.Run("keep old details on error", func(t *testing.T) {
		details := []data.Detail{
			{
				Key:        "key",
				Value:      "value",
				ProfileID:  1,
				ID:         1,
				DetailType: data.AliasDetail,
			},
		}
		detailPage := NewDetailPage(detailModelStub{getAll: func(profileID int) ([]data.Detail, error) { return nil, errors.New("unknown error") }})
		detailPage.details = details
		err := detailPage.resetDetails()
		assert.NotNil(t, err)
		assert.Equal(t, details, detailPage.details)
	})
}

func TestUpdatePaneStyle(t *testing.T) {
	testcases := []struct {
		name              string
		activePane        detailPagePane
		currentStage      detailStage
		displayStyle      lipgloss.Style
		actionsStyle      lipgloss.Style
		aliasStyle        lipgloss.Style
		envStyle          lipgloss.Style
		keyDisplayStyle   lipgloss.Style
		valueDisplayStyle lipgloss.Style
		confirmButton     lipgloss.Style
		cancelButton      lipgloss.Style
		deleteButton      lipgloss.Style
	}{}
}
