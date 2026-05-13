package main

import (
	"fmt"
	"os"

	"github.com/kbliu/review/internal/ui"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "review",
		Usage:     "Interactive terminal diff viewer",
		ArgsUsage: "[target]",
		Description: `Browse and review git diffs in an interactive TUI.

TARGET accepts any ref supported by 'git diff':

  review                   working tree vs HEAD (default)
  review HEAD              working tree vs HEAD
  review HEAD~3            working tree vs 3 commits ago
  review main              working tree vs branch main
  review abc1234           working tree vs a commit
  review main..feature     changes between two branches
  review HEAD~3..HEAD      diff of the last 3 commits
  review --staged          staged changes (index vs HEAD)`,
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "staged",
				Aliases: []string{"s"},
				Usage:   "Review staged changes (index vs HEAD)",
			},
			&cli.IntFlag{
				Name:    "unified",
				Aliases: []string{"U"},
				Value:   3,
				Usage:   "Number of context lines around each change",
			},
		},
		Action: func(c *cli.Context) error {
			target := c.Args().First()
			if target == "" && !c.Bool("staged") {
				target = "HEAD"
			}

			opts := ui.Options{
				Target:       target,
				Staged:       c.Bool("staged"),
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
