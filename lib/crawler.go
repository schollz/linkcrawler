package crawler

import (
	"bytes"
	"compress/gzip"
	"encoding/base32"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/goware/urlx"
	"github.com/jackdanger/collectlinks"
	"github.com/schollz/jsonstore"
)

// Crawler is the crawler instance
type Crawler struct {
	client               *http.Client
	wg                   sync.WaitGroup
	done                 *jsonstore.JSONStore
	todo                 *jsonstore.JSONStore
	trash                *jsonstore.JSONStore
	programTime          time.Time
	curFileList          map[string]bool
	numberOfURLSParsed   int
	BaseURL              string
	KeywordsToExclude    []string
	KeywordsToInclude    []string
	MaxNumberWorkers     int
	MaxNumberConnections int
	IterationsEverySave  int
	Verbose              bool
	FilePrefix           string
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
	c.IterationsEverySave = 5
	return c, nil
}

func (c *Crawler) loadKeyStores(urls []string, download bool) error {
	var err, err2 error

	// Initialize the keystores
	c.done = new(jsonstore.JSONStore)
	c.todo = new(jsonstore.JSONStore)
	c.trash = new(jsonstore.JSONStore)

	// Determine the file prefix
	filePrefix := c.FilePrefix
	if download {
		filePrefix = filePrefix + "_dl"
	} else {
		filePrefix = filePrefix + "_crawl"
	}

	// Load the keystores if they already exist
	if _, err = os.Stat(filePrefix + "_todo.json"); err == nil {
		c.todo, err2 = jsonstore.Open(filePrefix + "_todo.json")
		if err2 != nil {
			return err2
		}
	} else {
		for _, url := range urls {
			c.todo.Set(url, 0)
		}
	}

	if _, err = os.Stat(filePrefix + "_done.json"); err == nil {
		c.done, err2 = jsonstore.Open(filePrefix + "_done.json")
		if err2 != nil {
			return err2
		}
	}
	if _, err = os.Stat(filePrefix + "_trash.json"); err == nil {
		c.trash, err2 = jsonstore.Open(filePrefix + "_trash.json")
		if err2 != nil {
			return err2
		}
	}
	return nil
}

// saveData saves the data to the respective files
func (c *Crawler) saveKeyStores(download bool) (int, error) {
	// Determine the file prefix
	filePrefix := c.FilePrefix
	if download {
		filePrefix = filePrefix + "_dl"
	} else {
		filePrefix = filePrefix + "_crawl"
	}

	tempPrefix := randStringBytesMaskImprSrc(10)
	err1 := jsonstore.Save(c.done, tempPrefix+"_done.json")
	defer os.Remove(tempPrefix + "_done.json")
	err2 := jsonstore.Save(c.todo, tempPrefix+"_todo.json")
	defer os.Remove(tempPrefix + "_todo.json")
	err3 := jsonstore.Save(c.trash, tempPrefix+"_trash.json")
	defer os.Remove(tempPrefix + "_trash.json")
	if err1 != nil {
		return -1, err1
	} else if err2 != nil {
		return -1, err2
	} else if err3 != nil {
		return -1, err3
	}
	startSave := time.Now()
	copyFileContents(tempPrefix+"_done.json", filePrefix+"_done.json")
	copyFileContents(tempPrefix+"_todo.json", filePrefix+"_todo.json")
	copyFileContents(tempPrefix+"_trash.json", filePrefix+"_trash.json")
	log.Printf("Saved state in %s", time.Since(startSave).String())
	parsingTime := time.Since(c.programTime)
	URLperSecond := int(float64(c.numberOfURLSParsed) / parsingTime.Seconds())
	numTodo := len(c.todo.GetAll(regexp.MustCompile(`.*`)))
	numFinished := len(c.done.GetAll(regexp.MustCompile(`.*`)))
	log.Printf("Parsed %s urls in %s (%d URLs/s). Finished: %s, Todo: %s\n",
		humanize.Comma(int64(c.numberOfURLSParsed)),
		parsingTime.String(),
		URLperSecond,
		humanize.Comma(int64(numFinished)),
		humanize.Comma(int64(numTodo)))
	return numTodo, nil
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
			c.todo.Delete(url)
			c.done.Set(url, 0)
			return nil
		}
	}

	var foo, currentNumberOfTries int

	// Skip if the URL is done or trashed
	if c.done.Get(url, &foo) == nil || c.trash.Get(url, &foo) == nil {
		c.todo.Delete(url)
		return nil
	}

	c.todo.Get(url, &currentNumberOfTries)
	currentNumberOfTries++

	resp, err := c.client.Get(url)
	if err != nil {
		c.trash.Set(url, currentNumberOfTries)
		c.todo.Delete(url)
		if c.Verbose {
			log.Printf("Problem with %s: %s", url, err.Error())
		}
		return err
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
				if c.todo.Get(normalizedLink, &foo) != nil &&
					c.done.Get(normalizedLink, &foo) != nil &&
					c.trash.Get(normalizedLink, &foo) != nil {
					c.todo.Set(normalizedLink, 0)
				}
			}
		}

		// Dequeue the current URL
		c.done.Set(url, currentNumberOfTries)
		c.todo.Delete(url)
	} else {
		if currentNumberOfTries > 3 {
			// Delete this URL as it has been tried too many times
			c.trash.Set(url, currentNumberOfTries)
			c.todo.Delete(url)
			if c.Verbose {
				log.Println("Too many tries, trashing " + url)
			}
		} else {
			c.todo.Set(url, currentNumberOfTries)
		}
	}
	return nil
}

// Get a list of the done urls
func (c *Crawler) GetLinks() []string {
	doneMap := c.done.GetAll(regexp.MustCompile(`.*`))
	todoMap := c.todo.GetAll(regexp.MustCompile(`.*`))
	allLinks := make([]string, len(doneMap)+len(todoMap))
	i := 0
	for link := range doneMap {
		allLinks[i] = link
		i++
	}
	for link := range todoMap {
		allLinks[i] = link
		i++
	}
	return allLinks
}

// Crawl downloads the pages specified in the todo file
func (c *Crawler) Download(urls []string) error {
	download := true

	// Initialize the keystores
	err := c.loadKeyStores(urls, download)
	if err != nil {
		return err
	}

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

	return c.downloadOrCrawl(download)
}

// Crawl is the function to crawl with the set parameters
func (c *Crawler) Crawl() error {
	download := false

	// Initialize/load the key stores
	err := c.loadKeyStores([]string{c.BaseURL}, download)
	if err != nil {
		return err
	}

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
	for {
		it++
		numWorkers := 0
		for url := range c.todo.GetAll(regexp.MustCompile(`.*`)) {
			numWorkers++
			if numWorkers > c.MaxNumberWorkers {
				break
			}
			c.wg.Add(1)
			go c.collectLinks(url, download)
		}
		c.wg.Wait()

		// Save every 5th iteration
		if math.Mod(float64(it), float64(c.IterationsEverySave)) == 0 {
			numTodo, err2 := c.saveKeyStores(download)
			if err2 != nil || numTodo == 0 {
				return err2
			}
		}
	}
}

// copyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func copyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
