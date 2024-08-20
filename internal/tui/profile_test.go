package tui

import (
	"errors"
	"testing"

	"github.com/bento01dev/maggi/internal/data"
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
