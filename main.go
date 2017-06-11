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

var version string

func main() {
	app := cli.NewApp()
	app.Name = "linkcrawler"
	app.Usage = "crawl a site for links, or download a list of sites"
	app.Version = version
	app.Compiled = time.Now()
	app.Action = func(c *cli.Context) error {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "server, s",
			Value: "",
			Usage: "boltdb server instance [required]",
		},
		cli.StringFlag{
			Name:  "useragent,",
			Value: "",
			Usage: "supply a User-Agent string to be used",
		},
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
			Name:  "stats",
			Value: 1,
			Usage: "Print stats every `X` seconds",
		},
		cli.IntFlag{
			Name:  "trash-limit",
			Value: 5,
			Usage: "Exit if trashed URLs accumulates more than `X` / stats check",
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
		cli.BoolFlag{
			Name:  "redo",
			Usage: "move doing to todo",
		},
		cli.BoolFlag{
			Name:  "tor",
			Usage: "use tor proxy",
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

				if c.GlobalString("server") == "" {
					fmt.Println("Must specify BoltDB server ")
					return nil
				}

				fmt.Println(url)

				// Setup crawler to crawl
				fmt.Println("Setting up crawler...")
				craw, err := crawler.New(url, c.GlobalString("server"), c.GlobalBool("verbose"))
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
				craw.UserAgent = c.GlobalString("useragent")
				craw.TrashLimit = c.GlobalInt("trash-limit")
				craw.UseTor = c.GlobalBool("tor")
				if len(c.GlobalString("include")) > 0 {
					craw.KeywordsToInclude = strings.Split(strings.ToLower(c.GlobalString("include")), ",")
				}
				if len(c.GlobalString("exclude")) > 0 {
					craw.KeywordsToExclude = strings.Split(strings.ToLower(c.GlobalString("exclude")), ",")
				}
				if c.GlobalBool("redo") {
					craw.ResetDoing()
				}
				fmt.Printf("Starting crawl using DB %s\n", craw.Name())
				err = craw.Crawl()
				if err != nil {
					return err
				}
				return craw.Dump()
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

				if c.GlobalString("server") == "" {
					fmt.Println("Must specify BoltDB server ")
					return nil
				}

				b, err := ioutil.ReadFile(fileWithListOfURLS)
				if err != nil {
					return err
				}
				links := strings.Split(string(b), "\n")

				// Setup crawler to download
				craw, err := crawler.New(fileWithListOfURLS, c.GlobalString("server"), c.GlobalBool("verbose"))
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
				craw.UserAgent = c.GlobalString("useragent")
				craw.TrashLimit = c.GlobalInt("trash-limit")
				craw.UseTor = c.GlobalBool("tor")
				if len(c.GlobalString("include")) > 0 {
					craw.KeywordsToInclude = strings.Split(strings.ToLower(c.GlobalString("include")), ",")
				}
				if len(c.GlobalString("exclude")) > 0 {
					craw.KeywordsToExclude = strings.Split(strings.ToLower(c.GlobalString("exclude")), ",")
				}
				if c.GlobalBool("redo") {
					craw.ResetDoing()
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
				url := ""
				if c.NArg() > 0 {
					url = c.Args().Get(0)
				} else {
					fmt.Println("Must specify url to dump")
					return nil
				}

				if c.GlobalString("server") == "" {
					fmt.Println("Must specify BoltDB server ")
					return nil
				}

				// Setup crawler to crawl
				craw, err := crawler.New(url, c.GlobalString("server"), c.GlobalBool("verbose"))
				if err != nil {
					return err
				}
				return craw.Dump()
			},
		},
	}

	app.Run(os.Args)

}
