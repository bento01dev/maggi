package main

import (
	"log"
	"os"

	"github.com/bento01dev/maggi/internal/generate"
	"github.com/bento01dev/maggi/internal/hooks"
	"github.com/bento01dev/maggi/internal/tui"
	"github.com/urfave/cli/v2"
)

func main() {
	runApp()
}

func runApp() {
	var profileStr string
	var debugFlag bool
	var hookType string
	var globalHook bool

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
					return tui.Run(debugFlag)
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
					return generate.Run(profileStr)
				},
			},
			{
				Name:  "hook",
				Usage: "create, delete hooks to invoke maggi",
				Subcommands: []*cli.Command{
					{
						Name:  "add",
						Usage: "add hooks",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "type",
								Usage:       "type of hook to be added. currently only suuports tmux",
								Destination: &hookType,
								Required:    true,
							},
							&cli.BoolFlag{
								Name:        "global",
								Usage:       "hook applied globally",
								Destination: &globalHook,
								Value:       true,
							},
						},
						Action: func(ctx *cli.Context) error {
							return hooks.AddHook(hookType, globalHook)
						},
					},
					{
						Name:  "delete",
						Usage: "delete hooks",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "type",
								Usage:       "type of hook to be added. currently only suuports tmux",
								Destination: &hookType,
								Required:    true,
							},
							&cli.BoolFlag{
								Name:        "global",
								Usage:       "hook applied globally",
								Destination: &globalHook,
								Value:       true,
							},
						},
						Action: func(ctx *cli.Context) error {
							return hooks.DeleteHook(hookType, globalHook)
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
