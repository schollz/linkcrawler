package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/schollz/crawler/lib"
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

				// Setup crawler
				crawler, err := crawler.New(url)
				if err != nil {
					return err
				}
				if c.GlobalString("prefix") != "" {
					crawler.FilePrefix = c.GlobalString("prefix")
				}
				crawler.MaxNumberConnections = c.GlobalInt("conn")
				crawler.MaxNumberWorkers = c.GlobalInt("workers")
				crawler.Verbose = c.GlobalBool("verbose")
				if len(c.GlobalString("include")) > 0 {
					crawler.KeywordsToInclude = strings.Split(strings.ToLower(c.GlobalString("include")), ",")
				}
				if len(c.GlobalString("exclude")) > 0 {
					crawler.KeywordsToExclude = strings.Split(strings.ToLower(c.GlobalString("exclude")), ",")
				}
				err = crawler.Crawl()
				if err != nil {
					return err
				}
				linkArray := crawler.GetLinks()
				links := strings.Join(linkArray, "\n")
				ioutil.WriteFile("links.txt", []byte(links), 0755)
				fmt.Printf("%d links written to links.txt", len(linkArray))
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

				// Setup crawler
				crawler, err := crawler.New(fileWithListOfURLS)
				if err != nil {
					return err
				}
				if c.GlobalString("prefix") != "" {
					crawler.FilePrefix = c.GlobalString("prefix")
				}
				crawler.MaxNumberConnections = c.GlobalInt("conn")
				crawler.MaxNumberWorkers = c.GlobalInt("workers")
				crawler.Verbose = c.GlobalBool("verbose")
				if len(c.GlobalString("include")) > 0 {
					crawler.KeywordsToInclude = strings.Split(strings.ToLower(c.GlobalString("include")), ",")
				}
				if len(c.GlobalString("exclude")) > 0 {
					crawler.KeywordsToExclude = strings.Split(strings.ToLower(c.GlobalString("exclude")), ",")
				}

				b, err := ioutil.ReadFile(fileWithListOfURLS)
				if err != nil {
					return err
				}
				links := strings.Split(string(b), "\n")
				err = crawler.Download(links)
				if err != nil {
					return err
				}
				return nil
			},
		},
	}

	app.Run(os.Args)

}
