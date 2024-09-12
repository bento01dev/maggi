package generate

import (
	"testing"

	"github.com/bento01dev/maggi/internal/data"
	"github.com/stretchr/testify/assert"
)

type profileRepositoryStub struct {
	details []data.Detail
	err     error
}

func (p profileRepositoryStub) GetDetailsByProfileName(profileName string) ([]data.Detail, error) {
	return p.details, p.err
}

func TestGenerate(t *testing.T) {
	testcases := []struct {
		name        string
		details     []data.Detail
		err         error
		res         string
	}{
        {
            name: "empty details should return empty string",
        },
        {
            name: "non-empty details should return string with exports and aliases",
            details: []data.Detail{{Key: "test_env", Value: "test_env_value", DetailType: data.EnvDetail}, {Key: "test_alias", Value: "test_alias_value", DetailType: data.AliasDetail}},
            res: "export test_env=test_env_value;alias test_alias=test_alias_value;",
        },
    }

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			res, err := generate(profileRepositoryStub{testcase.details, testcase.err}, "")
			assert.Equal(t, testcase.res, res)
			assert.Equal(t, testcase.err, err)
		})
	}
}
