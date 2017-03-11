package crawler

import (
	"bytes"
	"compress/gzip"
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	humanize "github.com/dustin/go-humanize"
	"github.com/goware/urlx"
	"github.com/jackdanger/collectlinks"
	"github.com/schollz/archiver"
)

// Crawler is the crawler instance
type Crawler struct {
	client                   *http.Client
	wg                       sync.WaitGroup
	programTime              time.Time
	curFileList              map[string]bool
	BaseURL                  string
	KeywordsToExclude        []string
	KeywordsToInclude        []string
	MaxNumberWorkers         int
	MaxNumberConnections     int
	Verbose                  bool
	FilePrefix               string
	TimeIntervalToPrintStats int
	TimeIntervalToBackupDB   int
	numTrash                 int
	numDone                  int
	numToDo                  int
	numberOfURLSParsed       int
}

func encodeURL(url string) string {
	return base32.StdEncoding.EncodeToString([]byte(url))
}

// New will create a new crawler
func New(url string) (*Crawler, error) {
	c := new(Crawler)
	c.BaseURL = url
	c.MaxNumberConnections = 100
	c.MaxNumberWorkers = 100
	c.FilePrefix = encodeURL(url)
	c.TimeIntervalToPrintStats = 5
	c.TimeIntervalToBackupDB = 5
	c.numDone = 0
	c.numToDo = 0
	c.numTrash = 0
	return c, c.initDB()
}

func (c *Crawler) collectLinks(url string, download bool) error {
	// Decrement the counter when the goroutine completes.
	defer c.wg.Done()

	if download {
		// Check if it is already downloaded
		if _, ok := c.curFileList[encodeURL(url)]; ok {
			if c.Verbose {
				log.Printf("Already downloaded %s", url)
			}
			c.delete("todo", url)
			c.set("done", url, 0)
			return nil
		}
	}

	// Skip if the URL is done or trashed
	isDone, err := c.contains("done", url)
	if err != nil {
		return err
	}
	isTrashed, err := c.contains("trash", url)
	if err != nil {
		return err
	}
	if isDone || isTrashed {
		c.delete("todo", url)
		return nil
	}

	// Get the current number of tries
	currentNumberOfTries, err := c.get("todo", url)
	if err != nil {
		return err
	}
	currentNumberOfTries++

	resp, err := c.client.Get(url)
	if err != nil {
		err2 := c.set("trash", url, currentNumberOfTries)
		if err2 != nil {
			return err
		}
		err2 = c.delete("todo", url)
		if err2 != nil {
			return err
		}
		if c.Verbose {
			log.Printf("Problem with %s: %s", url, err.Error())
		}
		return nil
	}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		c.numberOfURLSParsed++

		// Download, if downloading
		if download {
			contentType := resp.Header.Get("Content-type")
			contentTypes, contentTypeErr := mime.ExtensionsByType(contentType)
			extension := ""
			if contentTypeErr == nil {
				extension = contentTypes[0]
				if extension == ".htm" || extension == ".hxa" {
					extension = ".html"
				}
			} else {
				return err
			}
			file_content, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			writer := gzip.NewWriter(&buf)
			writer.Write(file_content)
			writer.Close()
			filename := encodeURL(url) + extension + ".gz"
			os.Mkdir("downloaded", 0755)
			err = ioutil.WriteFile(path.Join("downloaded", filename), buf.Bytes(), 0755)
			if err != nil {
				return err
			}

			if c.Verbose {
				log.Printf("Saved %s to %s", url, encodeURL(url)+extension)
			}
		} else {
			links := collectlinks.All(resp.Body)
			if c.Verbose {
				log.Printf("Got %d links from %s\n", len(links), url)
			}
			for _, link := range links {
				// Do not use query parameters
				if strings.Contains(link, "?") {
					link = strings.Split(link, "?")[0]
				}
				// Add the Base URL to everything if it doesn't have it
				if !strings.Contains(link, "http") {
					link = c.BaseURL + link
				}
				// Skip links that have a different Base URL
				if !strings.Contains(link, c.BaseURL) {
					if c.Verbose {
						log.Printf("Skipping %s because it has a different base URL", link)
					}
					continue
				}
				// Normalize the link
				parsedLink, _ := urlx.Parse(link)
				normalizedLink, _ := urlx.Normalize(parsedLink)
				if len(normalizedLink) == 0 {
					continue
				}

				// Exclude keywords, skip if any are found
				foundExcludedKeyword := false
				for _, keyword := range c.KeywordsToExclude {
					if strings.Contains(normalizedLink, keyword) {
						foundExcludedKeyword = true
						if c.Verbose {
							log.Printf("Skipping %s because contains %s", link, keyword)
						}
						break
					}
				}
				if foundExcludedKeyword {
					continue
				}

				// Include keywords, skip if any are NOT found
				foundIncludedKeyword := false
				for _, keyword := range c.KeywordsToInclude {
					if strings.Contains(normalizedLink, keyword) {
						foundIncludedKeyword = true
						break
					}
				}
				if !foundIncludedKeyword && len(c.KeywordsToInclude) > 0 {
					continue
				}

				// Add the new link if its not queued anywhere else
				isDone, err := c.contains("done", normalizedLink)
				if err != nil {
					return err
				}
				isTrashed, err := c.contains("trash", normalizedLink)
				if err != nil {
					return err
				}
				isTodo, err := c.contains("todo", normalizedLink)
				if err != nil {
					return err
				}
				if !isDone && !isTrashed && !isTodo {
					c.set("todo", normalizedLink, 0)
					c.numToDo++
				}
			}
		}

		// Dequeue the current URL
		c.set("done", url, currentNumberOfTries)
		c.numDone++
		c.delete("todo", url)
		c.numToDo--
	} else {
		if currentNumberOfTries > 3 {
			// Delete this URL as it has been tried too many times
			c.set("trash", url, currentNumberOfTries)
			c.numTrash++
			c.delete("todo", url)
			c.numToDo--
			if c.Verbose {
				log.Println("Too many tries, trashing " + url)
			}
		} else {
			// Update the URL with the number of tries
			c.set("todo", url, currentNumberOfTries)
		}
	}
	return nil
}

// Crawl downloads the pages specified in the todo file
func (c *Crawler) Download(urls []string) error {
	download := true

	// Determine which files have been downloaded
	c.curFileList = make(map[string]bool)
	files, err := ioutil.ReadDir("downloaded")
	if err == nil {
		for _, f := range files {
			name := strings.Split(f.Name(), ".")[0]
			if len(name) < 2 {
				continue
			}
			c.curFileList[name] = true
		}
	}

	for _, url := range urls {
		c.set("todo", url, 0)
	}

	return c.downloadOrCrawl(download)
}

// Crawl is the function to crawl with the set parameters
func (c *Crawler) Crawl() error {
	c.set("todo", c.BaseURL, 0)
	download := false
	return c.downloadOrCrawl(download)
}

func (c *Crawler) downloadOrCrawl(download bool) error {
	// Generate the connection pool
	tr := &http.Transport{
		MaxIdleConns:       c.MaxNumberConnections,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	c.client = &http.Client{Transport: tr}

	c.programTime = time.Now()
	c.numberOfURLSParsed = 0
	it := 0
	go c.contantlyPrintStats()
	go c.contantlyPerformBackup()
	for {
		it++
		linksToDo, err := c.getNLinksTodo(c.MaxNumberWorkers)
		if err != nil {
			return err
		}
		if len(linksToDo) == 0 {
			break
		}
		for _, url := range linksToDo {
			c.wg.Add(1)
			go c.collectLinks(url, download)
		}
		c.wg.Wait()
	}
	c.numToDo = 0
	c.printStats()
	return nil
}

func round(f float64) int {
	if math.Abs(f) < 0.5 {
		return 0
	}
	return int(f + math.Copysign(0.5, f))
}

func (c *Crawler) contantlyPrintStats() {
	for {
		time.Sleep(time.Duration(int32(c.TimeIntervalToPrintStats)) * time.Second)
		if c.numToDo == 0 {
			fmt.Println("Finished")
			return
		}
		c.printStats()
	}
}

func (c *Crawler) printStats() {
	URLSPerSecond := round(float64(c.numberOfURLSParsed) / float64(time.Since(c.programTime).Seconds()))
	log.Printf("%s parsed (%d/s), %s todo, %s done, %s trashed\n",
		humanize.Comma(int64(c.numberOfURLSParsed)),
		URLSPerSecond,
		humanize.Comma(int64(c.numToDo)),
		humanize.Comma(int64(c.numDone)),
		humanize.Comma(int64(c.numTrash)))
}

func (c *Crawler) contantlyPerformBackup() {
	for {
		time.Sleep(time.Duration(int32(c.TimeIntervalToBackupDB)) * time.Second)
		os.Remove(c.FilePrefix + ".db.zip")
		db, _ := bolt.Open(c.FilePrefix+".db", 0600, &bolt.Options{Timeout: 10 * time.Second})
		err := archiver.Zip.Make(c.FilePrefix+".db.zip", []string{c.FilePrefix + ".db"})
		db.Close()
		if err == nil {
			fmt.Printf("%s\tBacked up DB\n", c.programTime.String())
		} else {
			fmt.Printf("%s\nError backing up DB:%s\n", c.programTime.String(), err.Error())
		}
	}
}
