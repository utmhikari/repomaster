package main

import (
	"github.com/urfave/cli/v2"
	"github.com/utmhikari/repomaster/internal"
	"os"
	"sort"
)

func main() {
	cliApp := cli.App{
		Name:  "repomaster",
		Usage: "master of repo",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "configs/repomaster.json",
				Usage:   "Load config from json file in configs dir",
			},
		},
		Action: func(c *cli.Context) error {
			cfgPath := c.String("config")
			return app.Start(cfgPath)
		},
	}

	sort.Sort(cli.FlagsByName(cliApp.Flags))
	sort.Sort(cli.CommandsByName(cliApp.Commands))

	err := cliApp.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
