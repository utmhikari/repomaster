package main

import (
	"github.com/urfave/cli/v2"
	"github.com/utmhikari/repomaster/internal"
	"github.com/utmhikari/repomaster/pkg/util"
	"os"
	"path"
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
				Value:   "repomaster",
				Usage:   "Load config from json file in configs dir",
			},
		},
		Action: func(c *cli.Context) error {
			cfgPath := path.Join("configs", c.String("config")+".json")
			appConfig := app.Config{
				Port: 8000,
			}
			err := util.ReadJsonFile(cfgPath, &appConfig)
			if err != nil {
				return err
			}
			return app.Start(appConfig)
		},
	}

	sort.Sort(cli.FlagsByName(cliApp.Flags))
	sort.Sort(cli.CommandsByName(cliApp.Commands))

	err := cliApp.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
