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
	}{
		{
			name:              "env pane is green when active",
			activePane:        envPane,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(muted),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(muted),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(green),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(muted),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(muted),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
		},
		{
			name:              "alias pane is green when active",
			activePane:        aliasPane,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(muted),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(muted),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(green),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(muted),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(muted),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
		},
		{
			name:              "display pane is green when active",
			activePane:        detailDisplayPane,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(green),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(muted),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(muted),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(muted),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
		},
		{
			name:              "action pane is green when active",
			activePane:        detailActionPane,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(muted),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(green),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(muted),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(muted),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
		},
		{
			name:              "edit detail key is blue when in stage",
			currentStage:      editDetailKey,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(muted),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(muted),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(green),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(blue),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(muted),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
		},
		{
			name:              "edit detail value is blue when in stage",
			currentStage:      editDetailValue,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(muted),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(muted),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(green),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(muted),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(blue),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
		},
		{
			name:              "edit detail confirm is highlighted when in stage",
			currentStage:      editDetailConfirm,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(muted),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(muted),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(green),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(muted),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(muted),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(green),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
		},
		{
			name:              "edit detail cancel is highlighted when in stage",
			currentStage:      editDetailCancel,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(muted),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(muted),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(green),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(muted),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(muted),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(green),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
		},
		{
			name:              "delete button should be red when in stage",
			currentStage:      deleteDetailConfirm,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(muted),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(muted),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(green),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(muted),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(muted),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(red),
		},
		{
			name:              "delete button cancel should be highlighted when in stage",
			currentStage:      deleteDetailCancel,
			displayStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultDisplayWidth).Height(defaultDisplayHeight).UnsetPadding().BorderForeground(muted),
			actionsStyle:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(defaultDisplayWidth).UnsetPadding().BorderForeground(muted),
			aliasStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(muted),
			envStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).Width(defaultSideBarWidth).UnsetPadding().BorderForeground(green),
			keyDisplayStyle:   lipgloss.NewStyle().Foreground(blue).PaddingTop((defaultDisplayHeight / 2) - 2).PaddingLeft(4).PaddingRight(1).Foreground(muted),
			valueDisplayStyle: lipgloss.NewStyle().Foreground(blue).PaddingLeft(4).PaddingTop(1).Foreground(muted),
			confirmButton:     lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
			cancelButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(green),
			deleteButton:      lipgloss.NewStyle().Padding(buttonPaddingVertical, buttonPaddingHorizontal).MarginLeft(1).Foreground(lipgloss.Color("0")).Background(muted),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			detailPage := NewDetailPage(detailModelStub{})
			detailPage.activePane = testcase.activePane
			detailPage.currentStage = testcase.currentStage
			detailPage.updatePaneStyles()
			assert.Equal(t, testcase.displayStyle, detailPage.displayStyle)
			assert.Equal(t, testcase.actionsStyle, detailPage.actionsStyle)
			assert.Equal(t, testcase.aliasStyle, detailPage.aliasStyle)
			assert.Equal(t, testcase.envStyle, detailPage.envStyle)
			assert.Equal(t, testcase.keyDisplayStyle, detailPage.keyDisplayStyle)
			assert.Equal(t, testcase.valueDisplayStyle, detailPage.valueDisplayStyle)
			assert.Equal(t, testcase.confirmButton, detailPage.confirmButton)
			assert.Equal(t, testcase.cancelButton, detailPage.cancelButton)
			assert.Equal(t, testcase.deleteButton, detailPage.deleteButton)
		})
	}
}

func TestDetailResetInfoBag(t *testing.T) {
	t.Run("all info bag fields should be reset on invocation", func(t *testing.T) {
		detailPage := NewDetailPage(detailModelStub{})
		detailPage.infoMsg = "test"
		detailPage.infoFlag = true
		detailPage.isErrInfo = true

		detailPage.resetInfoBag()

		assert.False(t, detailPage.infoFlag)
		assert.False(t, detailPage.isErrInfo)
		assert.Equal(t, "", detailPage.infoMsg)
	})
}

func TestCheckIfKeyExists(t *testing.T) {
	testcases := []struct {
		name            string
		key             string
		details         []data.Detail
		currentUserFlow detailsUserFlow
		currentDetail   *data.Detail
		expected        bool
	}{
		{
			name: "name doesnt exist in details for new detail",
			key:  "new_key",
			details: []data.Detail{
				{
					Key:        "key",
					Value:      "value",
					ProfileID:  1,
					ID:         1,
					DetailType: data.AliasDetail,
				},
			},
			currentUserFlow: newDetail,
		},
		{
			name: "name exist in details for new detail",
			key:  "key",
			details: []data.Detail{
				{
					Key:        "key",
					Value:      "value",
					ProfileID:  1,
					ID:         1,
					DetailType: data.AliasDetail,
				},
			},
			currentUserFlow: newDetail,
			expected:        true,
		},
		{
			name: "name exist in details for update detail and not the same detail",
			key:  "key",
			details: []data.Detail{
				{
					Key:        "key",
					Value:      "value",
					ProfileID:  1,
					ID:         1,
					DetailType: data.AliasDetail,
				},
			},
			currentUserFlow: updateDetail,
			currentDetail:   &data.Detail{ID: 2},
			expected:        true,
		},
		{
			name: "name exist in details for update detail but the same detail",
			key:  "key",
			details: []data.Detail{
				{
					Key:        "key",
					Value:      "value",
					ProfileID:  1,
					ID:         1,
					DetailType: data.AliasDetail,
				},
			},
			currentUserFlow: updateDetail,
			currentDetail:   &data.Detail{ID: 1},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			detailPage := NewDetailPage(detailModelStub{})
			detailPage.details = testcase.details
			detailPage.currentUserFlow = testcase.currentUserFlow
			detailPage.currentDetail = testcase.currentDetail

			res := detailPage.checkIfKeyExists(testcase.key)
			assert.Equal(t, testcase.expected, res)
		})
	}
}

func TestDetailHandleEsc(t *testing.T) {
	t.Run("handle Esc should set all the necessary fields to default values", func(t *testing.T) {
		detailPage := NewDetailPage(detailModelStub{})
		detailPage.currentUserFlow = viewDetail
		detailPage.activePane = detailDisplayPane
		detailPage.currentDetail = &data.Detail{ID: 1}
		detailPage.emptyDisplay = false
		detailPage.infoMsg = "test"
		detailPage.isErrInfo = true
		detailPage.infoFlag = true
		detailPage.keyInput.SetValue("test")
		detailPage.valueInput.SetValue("test")
		detailPage.keyTextArea.SetValue("test")
		detailPage.valueTextArea.SetValue("test")

		detailPage.handleEsc()

		assert.Equal(t, listDetails, detailPage.currentUserFlow)
		assert.Equal(t, envPane, detailPage.activePane)
		assert.Nil(t, detailPage.currentDetail)
		assert.True(t, detailPage.emptyDisplay)
		assert.Equal(t, "", detailPage.infoMsg)
		assert.False(t, detailPage.isErrInfo)
		assert.False(t, detailPage.infoFlag)
		assert.Equal(t, "", detailPage.keyInput.Value())
		assert.Equal(t, "", detailPage.valueInput.Value())
		assert.Equal(t, "", detailPage.keyTextArea.Value())
		assert.Equal(t, "", detailPage.valueTextArea.Value())
	})
}

func TestHandleListDetailsTab(t *testing.T) {
	testcases := []struct {
		name              string
		shift             bool
		emptyDisplay      bool
		currentActivePane detailPagePane
		activePane        detailPagePane
		currentDetailType detailType
		detailType        detailType
	}{
		{
			name:              "shift tab on env pane should move to alias pane and alias type",
			shift:             true,
			currentActivePane: envPane,
			activePane:        aliasPane,
			currentDetailType: detailTypeEnv,
			detailType:        detailTypeAlias,
		},
		{
			name:              "shift tab on alias pane should move to alias pane and alias type, if no display",
			shift:             true,
			emptyDisplay:      true,
			currentActivePane: aliasPane,
			activePane:        envPane,
			currentDetailType: detailTypeAlias,
			detailType:        detailTypeEnv,
		},
		{
			name:              "shift tab on alias pane should move to detail action pane, when there is a display",
			shift:             true,
			currentActivePane: aliasPane,
			activePane:        detailActionPane,
			currentDetailType: detailTypeAlias,
			detailType:        detailTypeAlias,
		},
		{
			name:              "shift tab on action pane should switch to display pane",
			shift:             true,
			currentActivePane: detailActionPane,
			activePane:        detailDisplayPane,
			currentDetailType: detailTypeAlias,
			detailType:        detailTypeAlias,
		},
		{
			name:              "shift tab on display pane should switch to env pane and detail type",
			shift:             true,
			currentActivePane: detailDisplayPane,
			activePane:        envPane,
			currentDetailType: detailTypeAlias,
			detailType:        detailTypeEnv,
		},
		{
			name:              "tab on env pane on empty display should switch to alias pane and detail type",
			emptyDisplay:      true,
			currentActivePane: envPane,
			activePane:        aliasPane,
			currentDetailType: detailTypeEnv,
			detailType:        detailTypeAlias,
		},
		{
			name:              "tab on env pane should switch to display pane when there is a display",
			currentActivePane: envPane,
			activePane:        detailDisplayPane,
			currentDetailType: detailTypeEnv,
			detailType:        detailTypeEnv,
		},
		{
			name:              "tab on alias pane should switch to env pane and detail type",
			currentActivePane: aliasPane,
			activePane:        envPane,
			currentDetailType: detailTypeAlias,
			detailType:        detailTypeEnv,
		},
		{
			name:              "tab on action pane should switch to alias pane and detail type",
			currentActivePane: detailActionPane,
			activePane:        aliasPane,
			currentDetailType: detailTypeEnv,
			detailType:        detailTypeAlias,
		},
		{
			name:              "tab on display pane should move to action pane",
			currentActivePane: detailDisplayPane,
			activePane:        detailActionPane,
			currentDetailType: detailTypeEnv,
			detailType:        detailTypeEnv,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			detailPage := NewDetailPage(detailModelStub{})
			detailPage.emptyDisplay = testcase.emptyDisplay
			detailPage.activePane = testcase.currentActivePane
			detailPage.detailType = testcase.currentDetailType

			detailPage.handleListDetailsTab(testcase.shift)

			assert.Equal(t, testcase.activePane, detailPage.activePane)
			assert.Equal(t, testcase.detailType, detailPage.detailType)
		})
	}
}
