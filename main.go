package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	runApp()
}

func runApp() {
	var profile string

	app := &cli.App{
		Version:              "0.1",
		Name:                 "maggi",
		Commands: []*cli.Command{
			{
				Name:  "ui",
				Usage: "UI for managing your aliases",
				Action: func(ctx *cli.Context) error {
                    fmt.Println("ui for maggi")
					return nil
				},
			},
			{
				Name:  "profiles",
				Usage: "List active profiles",
				Action: func(ctx *cli.Context) error {
                    fmt.Println("profile list")
					return nil
				},
			},
			{
				Name:  "generate",
				Usage: "generate the alias file for give profile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "profile",
						Value:       "dev",
						Usage:       "pass profile for generating the required alias file",
						Destination: &profile,
						Required:    true,
					},
				},
				Action: func(ctx *cli.Context) error {
                    fmt.Println("generate alias file")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
