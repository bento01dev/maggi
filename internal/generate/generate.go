package generate

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bento01dev/maggi/internal/data"
)

type GenerateProfileRepository interface {
	GetDetailsByProfileName(name string) ([]data.Detail, error)
}

func GenerateForProfile(profileName string) error {
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

func GenerateForSession(defaultProfile string) error {
	db, err := data.Setup()
	if err != nil {
		return err
	}
	defer db.Close()

	profileRepository := data.NewProfileRepository(db)
	var defaultDetails string
	var profileDetails string
	if defaultProfile != "" {
		defaultDetails, err = generate(profileRepository, defaultProfile)
		if err != nil {
			fmt.Println("")
			return err
		}
	}
	if tmuxEnv := os.Getenv("TMUX"); tmuxEnv != "" {
		out, err := exec.Command("tmux", "display-message", "-p", "'#S'").Output()
		if err != nil {
			fmt.Println("")
			return err
		}
		profileName := string(out)
		profileName = strings.TrimSpace(profileName)
		profileName = strings.Trim(profileName, "'")
		profileName = strings.Trim(profileName, "\"")
		profileDetails, err = generate(profileRepository, profileName)
		if err != nil {
			fmt.Println("")
			return err
		}
	}

	fmt.Printf("%s%s", defaultDetails, profileDetails)

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
