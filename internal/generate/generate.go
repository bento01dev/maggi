package generate

import (
	"fmt"
	"strings"

	"github.com/bento01dev/maggi/internal/data"
)

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

	details, err := profileRepository.GetDetailsByProfileName(profileName)
	if err != nil {
		return err
	}

	fmt.Println(generate(details))

	return nil
}

func generate(details []data.Detail) string {
	var b strings.Builder

	for _, detail := range details {
		switch detail.DetailType {
		case data.AliasDetail:
			fmt.Fprintf(&b, "alias %s=%s;", detail.Key, detail.Value)
		case data.EnvDetail:
			fmt.Fprintf(&b, "export %s=%s;", detail.Key, detail.Value)
		}
	}
	return b.String()
}
