package main

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/crawler/crawler"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "linkcrawler"
	app.Usage = "crawl a site for links, or download a list of sites"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Action = func(c *cli.Context) error {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "exclude, e",
			Value: "",
			Usage: "comma-delimted phrases that must NOT be in URL",
		},
		cli.StringFlag{
			Name:  "include, i",
			Value: "",
			Usage: "comma-delimted phrases that must be in URL",
		},
		cli.IntFlag{
			Name:  "workers,w",
			Value: 100,
			Usage: "Max number of workers",
		},
		cli.IntFlag{
			Name:  "conn,c",
			Value: 100,
			Usage: "Max number of connections in HTTP pool",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "turn on logging",
		},
		cli.StringFlag{
			Name:  "prefix, p",
			Value: "",
			Usage: "override file prefix",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "crawl",
			Aliases: []string{"c"},
			Usage:   "crawl a website and get a list of links",
			Action: func(c *cli.Context) error {
				url := ""
				if c.NArg() > 0 {
					url = c.Args().Get(0)
				} else {
					fmt.Println("Must specify url to crawl")
					return nil
				}
				fmt.Println(c.GlobalString("lang"))
				fmt.Println(url)
				crawl, err := New(url)
				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:    "download",
			Aliases: []string{"d"},
			Usage:   "download a list of websites",
			Action: func(c *cli.Context) error {
				fileWithListOfURLS := ""
				if c.NArg() > 0 {
					fileWithListOfURLS = c.Args().Get(0)
				} else {
					fmt.Println("Must specify file containing list of URLs")
					return nil
				}
				fmt.Println(c.GlobalString("lang"))
				fmt.Println(fileWithListOfURLS)
				return nil
			},
		},
	}

	app.Run(os.Args)
}
