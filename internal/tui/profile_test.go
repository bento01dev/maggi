package tui

import (
	"errors"
	"testing"

	"github.com/bento01dev/maggi/internal/data"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

type profileModelStub struct {
	getAll        func() ([]data.Profile, error)
	add           func(name string) (data.Profile, error)
	update        func(profile data.Profile, newName string) (data.Profile, error)
	deleteProfile func(profile data.Profile) error
}

func (ps profileModelStub) GetAll() ([]data.Profile, error) {
	return ps.getAll()
}

func (ps profileModelStub) Add(name string) (data.Profile, error) {
	return ps.add(name)
}

func (ps profileModelStub) Update(profile data.Profile, newName string) (data.Profile, error) {
	return ps.update(profile, newName)
}

func (ps profileModelStub) Delete(profile data.Profile) error {
	return ps.deleteProfile(profile)
}

func TestResetInfoBag(t *testing.T) {
	testcases := []struct {
		name      string
		infoFlag  bool
		isErrInfo bool
		infoMsg   string
	}{
		{
			name: "empty bag should stay empty",
		},
		{
			name:     "normal info should also be set back to empty",
			infoFlag: true,
			infoMsg:  "test msg",
		},
		{
			name:      "error info should be set back to empty",
			infoFlag:  true,
			isErrInfo: true,
			infoMsg:   "error msg",
		},
	}

	profilePage := NewProfilePage(nil)
	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			profilePage.infoFlag = test.infoFlag
			profilePage.isErrInfo = test.isErrInfo
			profilePage.infoMsg = test.infoMsg

			profilePage.resetInfoBag()

			assert.False(t, profilePage.infoFlag)
			assert.False(t, profilePage.isErrInfo)
			assert.Equal(t, profilePage.infoMsg, "")
		})
	}
}

func TestGetProfiles(t *testing.T) {
	testcases := []struct {
		name     string
		getAll   func() ([]data.Profile, error)
		err      error
		expected []data.Profile
	}{
		{
			name: "returns a normal set of profiles",
			getAll: func() ([]data.Profile, error) {
				return []data.Profile{
					{
						ID:   1,
						Name: "test1",
					},
					{
						ID:   2,
						Name: "test2",
					},
				}, nil
			},
			expected: []data.Profile{
				{
					ID:   1,
					Name: "test1",
				},
				{
					ID:   2,
					Name: "test2",
				},
			},
		},
		{
			name: "returns empty profiles when there are no profiles",
			getAll: func() ([]data.Profile, error) {
				return []data.Profile{}, nil
			},
			expected: []data.Profile{},
		},
		{
			name: "error in retrieving profiles",
			getAll: func() ([]data.Profile, error) {
				return nil, errors.New("sql: Scan called without calling Next")
			},
			expected: nil,
			err:      errors.New("sql: Scan called without calling Next"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			profilePage := NewProfilePage(profileModelStub{getAll: testcase.getAll})
			msg := profilePage.getProfiles()
			assert.Equal(t, msg.err, testcase.err)
			assert.Equal(t, msg.profiles, testcase.expected)
		})
	}
}

func TestAddProfile(t *testing.T) {
	testcases := []struct {
		name        string
		profileName string
		add         func(name string) (data.Profile, error)
		err         error
		pre         []data.Profile
		post        []data.Profile
	}{
		{
			name:        "successfully add a profile",
			profileName: "test",
			add:         func(name string) (data.Profile, error) { return data.Profile{ID: 1, Name: "test"}, nil },
			pre:         []data.Profile{},
			post:        []data.Profile{{ID: 1, Name: "test"}},
		},
		{
			name:        "error in adding profile",
			profileName: "test2",
			add: func(name string) (data.Profile, error) {
				return data.Profile{}, errors.New("no LastInsertId available after DDL statement")
			},
			pre:  []data.Profile{{ID: 1, Name: "test"}},
			post: []data.Profile{{ID: 1, Name: "test"}},
			err:  errors.New("no LastInsertId available after DDL statement"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			profilePage := NewProfilePage(profileModelStub{add: testcase.add})
			profilePage.profiles = testcase.pre
			err := profilePage.addProfile(testcase.profileName)
			assert.Equal(t, err, testcase.err)
			assert.Equal(t, profilePage.profiles, testcase.post)
		})
	}
}

func TestGetItemsMaxLen(t *testing.T) {
	testcases := []struct {
		name     string
		elems    []string
		expected int
	}{
		{
			name:     "standard slice of elements",
			elems:    []string{"global", "prd", "stg", "dev"},
			expected: 6,
		},
		{
			name: "returns 0 when there are no elements",
		},
	}

	profilePage := NewProfilePage(profileModelStub{})
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			res := profilePage.getItemsMaxLen(testcase.elems)
			assert.Equal(t, res, testcase.expected)
		})
	}
}

func TestCheckDuplicate(t *testing.T) {
	testcases := []struct {
		name        string
		profiles    []data.Profile
		profileName string
		expected    bool
	}{
		{
			name:        "name is not a duplicate",
			profiles:    []data.Profile{{ID: 1, Name: "test"}},
			profileName: "test1",
		},
		{
			name:        "name is duplicate",
			profiles:    []data.Profile{{ID: 1, Name: "test"}},
			profileName: "test",
			expected:    true,
		},
	}

	profilePage := NewProfilePage(profileModelStub{})
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			profilePage.profiles = testcase.profiles
			res := profilePage.checkDuplicate(testcase.profileName)
			assert.Equal(t, res, testcase.expected)
		})
	}
}

func TestResetProfiles(t *testing.T) {
	testcases := []struct {
		name             string
		err              error
		pre              []data.Profile
		post             []data.Profile
		getAll           func() ([]data.Profile, error)
		newProfileOption bool
	}{
		{
			name: "updates profiles list as per normal",
			pre:  []data.Profile{{ID: 1, Name: "test1"}},
			post: []data.Profile{{ID: 1, Name: "test1"}, {ID: 2, Name: "test2"}},
			getAll: func() ([]data.Profile, error) {
				return []data.Profile{{ID: 1, Name: "test1"}, {ID: 2, Name: "test2"}}, nil
			},
		},
		{
			name:             "profile list updated to empty list",
			pre:              []data.Profile{{ID: 1, Name: "test1"}},
			post:             []data.Profile{},
			getAll:           func() ([]data.Profile, error) { return []data.Profile{}, nil },
			newProfileOption: true,
		},
		{
			name: "dont update profile or new profile flag on error",
			pre:  []data.Profile{{ID: 1, Name: "test1"}},
			post: []data.Profile{{ID: 1, Name: "test1"}},
			getAll: func() ([]data.Profile, error) {
				return []data.Profile{}, errors.New("sql: Scan called without calling Next")
			},
			err: errors.New("sql: Scan called without calling Next"),
		},
		{
			name:             "empty profile update so set new profile option",
			pre:              []data.Profile{{ID: 1, Name: "test1"}},
			post:             []data.Profile{},
			getAll:           func() ([]data.Profile, error) { return []data.Profile{}, nil },
			newProfileOption: true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			profilePage := NewProfilePage(profileModelStub{getAll: testcase.getAll})
			profilePage.profiles = testcase.pre
			err := profilePage.resetProfiles()
			assert.Equal(t, err, testcase.err)
			assert.Equal(t, profilePage.profiles, testcase.post)
		})
	}
}

func TestUpdateActionStyle(t *testing.T) {
	testcases := []struct {
		name       string
		activePane profilePagePane
		expected   lipgloss.Color
	}{
		{
			name:       "muted when profile pane is active",
			activePane: profilesPane,
			expected:   muted,
		},
		{
			name:       "green when action pane is active",
			activePane: actionsPane,
			expected:   green,
		},
	}

	profilePage := NewProfilePage(profileModelStub{})
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			profilePage.activePane = testcase.activePane
			profilePage.updateActionStyle()
			assert.Equal(t, profilePage.actionsStyle.GetBorderBottomForeground(), testcase.expected)
			assert.Equal(t, profilePage.actionsStyle.GetBorderTopForeground(), testcase.expected)
			assert.Equal(t, profilePage.actionsStyle.GetBorderLeftForeground(), testcase.expected)
			assert.Equal(t, profilePage.actionsStyle.GetBorderRightForeground(), testcase.expected)
		})
	}
}

func TestUpdateProfileStyle(t *testing.T) {
	testcases := []struct {
		name       string
		activePane profilePagePane
		expected   lipgloss.Color
	}{
		{
			name:       "muted when action pane is active",
			activePane: actionsPane,
			expected:   muted,
		},
		{
			name:       "green when profile pane is active",
			activePane: profilesPane,
			expected:   green,
		},
	}

	profilePage := NewProfilePage(profileModelStub{})
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			profilePage.activePane = testcase.activePane
			profilePage.updateProfileStyle()
			assert.Equal(t, profilePage.profilesStyle.GetBorderBottomForeground(), testcase.expected)
			assert.Equal(t, profilePage.profilesStyle.GetBorderTopForeground(), testcase.expected)
			assert.Equal(t, profilePage.profilesStyle.GetBorderLeftForeground(), testcase.expected)
			assert.Equal(t, profilePage.profilesStyle.GetBorderRightForeground(), testcase.expected)
		})
	}
}

func TestHandleListProfilesTab(t *testing.T) {
	testcases := []struct {
		name string
		pre  profilePagePane
		post profilePagePane
	}{
		{
			name: "switch to action pane when on profiles",
			pre:  profilesPane,
			post: actionsPane,
		},
		{
			name: "switch to profile pane when on actions",
			pre:  actionsPane,
			post: profilesPane,
		},
	}

	profilePage := NewProfilePage(profileModelStub{})
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			profilePage.activePane = testcase.pre
			profilePage.handleListProfilesTab()
			assert.Equal(t, profilePage.activePane, testcase.post)
		})
	}
}

func TestHandleNewProfileTab(t *testing.T) {
	testcases := []struct {
		name          string
		shift         bool
		activePane    profilePagePane
		newPane       profilePagePane
		currentStage  profileStage
		newStage      profileStage
		preInfoFlag   bool
		preIsErrInfo  bool
		preInfoMsg    string
		postInfoFlag  bool
		postIsErrInfo bool
		postInfoMsg   string
	}{
		{
			name:       "shift tab on profile pane switches to actions pane",
			shift:      true,
			activePane: profilesPane,
			newPane:    actionsPane,
		},
		{
			name:         "shift tab on actions pane should switch to profiles and reset info if stage is add profile",
			shift:        true,
			activePane:   actionsPane,
			newPane:      profilesPane,
			currentStage: addProfileName,
			newStage:     addProfileName,
			preInfoFlag:  true,
			preIsErrInfo: true,
			preInfoMsg:   "test",
		},
		{
			name:         "shift tab on actions pane should switch stage to profile name when current one is profile confirm",
			shift:        true,
			activePane:   actionsPane,
			newPane:      actionsPane,
			currentStage: addProfileConfirm,
			newStage:     addProfileName,
		},
		{
			name:         "shift tab on actions pane should switch stage to profile confirm when current one is profile cancel",
			shift:        true,
			activePane:   actionsPane,
			newPane:      actionsPane,
			currentStage: addProfileCancel,
			newStage:     addProfileConfirm,
		},
		{
			name:       "tab on profile pane switches to actions pane",
			activePane: profilesPane,
			newPane:    actionsPane,
		},
		{
			name:         "tab on actions pane should change current stage from profile name to profile confirm",
			activePane:   actionsPane,
			newPane:      actionsPane,
			currentStage: addProfileName,
			newStage:     addProfileConfirm,
		},
		{
			name:         "tab on actions pane should change current stage from profile confirm to profile cancel",
			activePane:   actionsPane,
			newPane:      actionsPane,
			currentStage: addProfileConfirm,
			newStage:     addProfileCancel,
		},
		{
			name:         "tab on actions pane should switch to profiles and reset info if stage is add profile",
			activePane:   actionsPane,
			newPane:      profilesPane,
			currentStage: addProfileCancel,
			newStage:     addProfileName,
			preInfoFlag:  true,
			preIsErrInfo: true,
			preInfoMsg:   "test",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			profilePage := NewProfilePage(profileModelStub{})
			profilePage.activePane = testcase.activePane
			profilePage.currentStage = testcase.currentStage
			profilePage.infoFlag = testcase.preInfoFlag
			profilePage.isErrInfo = testcase.preIsErrInfo
			profilePage.infoMsg = testcase.preInfoMsg
			profilePage.handleNewProfileTab(testcase.shift)
			assert.Equal(t, profilePage.activePane, testcase.newPane)
			assert.Equal(t, profilePage.currentStage, testcase.newStage)
			assert.Equal(t, profilePage.infoFlag, testcase.postInfoFlag)
			assert.Equal(t, profilePage.isErrInfo, testcase.postIsErrInfo)
			assert.Equal(t, profilePage.infoMsg, testcase.postInfoMsg)
		})
	}
}
