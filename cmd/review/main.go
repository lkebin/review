package main

import (
	"fmt"
	"os"

	"github.com/kbliu/review/internal/ui"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "review",
		Usage:   "Code review tool for the terminal",
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "staged",
				Aliases: []string{"s"},
				Usage:   "Review staged changes",
			},
			&cli.IntFlag{
				Name:    "unified",
				Aliases: []string{"U"},
				Value:   3,
				Usage:   "Number of context lines",
			},
		},
		Action: func(c *cli.Context) error {
			target := c.Args().First()
			if target == "" && !c.Bool("staged") {
				target = "HEAD"
			}

			opts := ui.Options{
				Target:      target,
				Staged:      c.Bool("staged"),
				ContextLines: c.Int("unified"),
			}

			return ui.Run(opts)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
