package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/schollz/linkcrawler/lib"
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
		cli.IntFlag{
			Name:  "stats,s",
			Value: 1,
			Usage: "Print stats every `X` seconds",
		},
		cli.IntFlag{
			Name:  "backup,b",
			Value: 5,
			Usage: "Backup DB every `X` minutes",
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

				// Setup crawler to crawl
				craw, err := crawler.New(url)
				if err != nil {
					return err
				}
				if c.GlobalString("prefix") != "" {
					craw.FilePrefix = c.GlobalString("prefix")
				}
				craw.MaxNumberConnections = c.GlobalInt("conn")
				craw.MaxNumberWorkers = c.GlobalInt("workers")
				craw.Verbose = c.GlobalBool("verbose")
				craw.TimeIntervalToPrintStats = c.GlobalInt("stats")
				craw.TimeIntervalToBackupDB = c.GlobalInt("backup")
				if len(c.GlobalString("include")) > 0 {
					craw.KeywordsToInclude = strings.Split(strings.ToLower(c.GlobalString("include")), ",")
				}
				if len(c.GlobalString("exclude")) > 0 {
					craw.KeywordsToExclude = strings.Split(strings.ToLower(c.GlobalString("exclude")), ",")
				}
				err = craw.Crawl()
				if err != nil {
					return err
				}
				return crawler.Dump(craw.FilePrefix + ".db")
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

				b, err := ioutil.ReadFile(fileWithListOfURLS)
				if err != nil {
					return err
				}
				links := strings.Split(string(b), "\n")

				// Setup crawler to download
				craw, err := crawler.New(fileWithListOfURLS)
				if err != nil {
					return err
				}
				if c.GlobalString("prefix") != "" {
					craw.FilePrefix = c.GlobalString("prefix")
				}
				craw.MaxNumberConnections = c.GlobalInt("conn")
				craw.MaxNumberWorkers = c.GlobalInt("workers")
				craw.Verbose = c.GlobalBool("verbose")
				craw.TimeIntervalToPrintStats = c.GlobalInt("stats")
				craw.TimeIntervalToBackupDB = c.GlobalInt("backup")
				if len(c.GlobalString("include")) > 0 {
					craw.KeywordsToInclude = strings.Split(strings.ToLower(c.GlobalString("include")), ",")
				}
				if len(c.GlobalString("exclude")) > 0 {
					craw.KeywordsToExclude = strings.Split(strings.ToLower(c.GlobalString("exclude")), ",")
				}
				err = craw.Download(links)
				if err != nil {
					fmt.Printf("Error downloading: %s", err.Error())
					return err
				}
				fmt.Println("Finished downloading")
				return nil
			},
		},
		{
			Name:  "dump",
			Usage: "dump a list of links crawled from db",
			Action: func(c *cli.Context) error {
				dbFile := ""
				if c.NArg() > 0 {
					dbFile = c.Args().Get(0)
				} else {
					fmt.Println("Must specify database")
					return nil
				}
				return crawler.Dump(dbFile)
			},
		},
	}

	app.Run(os.Args)

}
