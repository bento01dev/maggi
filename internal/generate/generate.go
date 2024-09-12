package generate

import (
	"fmt"
	"strings"

	"github.com/bento01dev/maggi/internal/data"
)

type GenerateProfileRepository interface {
	GetDetailsByProfileName(name string) ([]data.Detail, error)
}

func Run(profileName string) error {
	//TODO: this should move to main once repository is used in tui
	db, err := data.Setup()
	if err != nil {
		return err
	}
	defer db.Close()

	profileRepository := data.NewProfileRepository(db)

	if profileName == "" {
		return nil
	}

	generatedStr, err := generate(profileRepository, profileName)
	if err != nil {
		return err
	}
	fmt.Println(generatedStr)

	return nil
}

func generate(repository GenerateProfileRepository, profileName string) (string, error) {
	details, err := repository.GetDetailsByProfileName(profileName)
	if err != nil {
		return "", err
	}
	var b strings.Builder

	for _, detail := range details {
		switch detail.DetailType {
		case data.AliasDetail:
			fmt.Fprintf(&b, "alias %s=%s;", detail.Key, detail.Value)
		case data.EnvDetail:
			fmt.Fprintf(&b, "export %s=%s;", detail.Key, detail.Value)
		}
	}
	return b.String(), nil
}
