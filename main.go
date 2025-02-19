package main

import (
	"log"
	"os"

	"github.com/bento01dev/maggi/internal/data"
	"github.com/bento01dev/maggi/internal/generate"
	"github.com/bento01dev/maggi/internal/tui"
	"github.com/urfave/cli/v2"
)

func main() {
	runApp()
}

func runApp() {
	var profileStr string
	var defaultProfile string
	var debugFlag bool

	app := &cli.App{
		Version: "0.1",
		Name:    "maggi",
		Commands: []*cli.Command{
			{
				Name:  "ui",
				Usage: "UI for managing your aliases",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "debug",
						Value:       false,
						Usage:       "debug mode",
						Destination: &debugFlag,
					},
				},
				Action: func(ctx *cli.Context) error {
					db, err := data.Setup()
					if err != nil {
						return err
					}
					defer db.Close()
					maggiRepository := data.NewMaggiRepository(db)
					return tui.Run(debugFlag, maggiRepository)
				},
			},
			{
				Name:  "generate",
				Usage: "generate the alias file for give profile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "profile",
						Usage:       "pass profile for generating the required alias file",
						Destination: &profileStr,
					},
				},
				Action: func(ctx *cli.Context) error {
					// should the error be dropped since the output is run via eval?
					db, err := data.Setup()
					if err != nil {
						return nil
					}
					defer db.Close()
					maggiRepository := data.NewMaggiRepository(db)
					generate.GenerateForProfile(profileStr, maggiRepository)
					return nil
				},
			},
			{
				Name:  "apply-session",
				Usage: "apply for a tmux session. can be set in .zprofile with defaults for regular non-tmux shell",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "default",
						Usage:       "default profile to apply. this alone will be applied when executed in non-tmux (regular) shell",
						Destination: &defaultProfile,
					},
				},
				Action: func(ctx *cli.Context) error {
					db, err := data.Setup()
					if err != nil {
						return nil
					}
					defer db.Close()
					maggiRepository := data.NewMaggiRepository(db)
					generate.GenerateForSession(defaultProfile, maggiRepository)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
