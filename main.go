package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base32"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"mime"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/goware/urlx"
	"github.com/jackdanger/collectlinks"
	"github.com/schollz/jsonstore"
)

func encodeURL(url string) string {
	return base32.StdEncoding.EncodeToString([]byte(url))
}

func crawl(baseURL string, keywordsToIgnore []string, keywordsToInclude []string, maxNumberWorkers int, connectionPool int, logging bool, download bool) {
	// Generate the connection pool
	var tr *http.Transport
	var client *http.Client
	tr = &http.Transport{
		MaxIdleConns:       connectionPool,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client = &http.Client{Transport: tr}

	var foo int
	var err error
	var wg sync.WaitGroup

	// Initialize the keystores
	doneURLS := new(jsonstore.JSONStore)
	todoURLS := new(jsonstore.JSONStore)
	trashURLS := new(jsonstore.JSONStore)
	// Load the keystores if they already exist
	if _, err = os.Stat("todoURLS.json"); err == nil {
		var err2 error
		todoURLS, err2 = jsonstore.Open("todoURLS.json")
		if err2 != nil {
			panic(err)
		}
	} else {
		todoURLS.Set(baseURL, true)
	}
	if _, err = os.Stat("doneURLS.json"); err == nil {
		var err2 error
		doneURLS, err2 = jsonstore.Open("doneURLS.json")
		if err2 != nil {
			panic(err)
		}
	}
	if _, err = os.Stat("trashURLS.json"); err == nil {
		var err2 error
		trashURLS, err2 = jsonstore.Open("trashURLS.json")
		if err2 != nil {
			panic(err)
		}
	}

	// Start scraping
	var curFileList map[string]bool
	if download {
		curFileList = make(map[string]bool)
	}
	programTime := time.Now()
	numberOfURLSParsed := 0
	it := 0
	for {
		it++
		if download {
			files, _ := ioutil.ReadDir("./")
			for _, f := range files {
				name := strings.Split(f.Name(), ".")[0]
				if len(name) < 2 {
					continue
				}
				curFileList[name] = true
			}
		}
		numWorkers := 0
		for URLToDo := range todoURLS.GetAll(regexp.MustCompile(`.*`)) {
			if download {
				if _, ok := curFileList[encodeURL(URLToDo)]; ok {
					if logging {
						log.Printf("Already downloaded %s", URLToDo)
					}
					todoURLS.Delete(URLToDo)
					doneURLS.Set(URLToDo, 0)
					continue
				}

			}
			numWorkers++
			if numWorkers > maxNumberWorkers {
				break
			}
			wg.Add(1)
			go func(url string) {
				// Decrement the counter when the goroutine completes.
				defer wg.Done()

				// Check if it is done
				if doneURLS.Get(url, &foo) == nil {
					todoURLS.Delete(url)
					return
				}

				var currentNumberOfTries int
				todoURLS.Get(url, &currentNumberOfTries)
				currentNumberOfTries++
				resp, err := client.Get(url)
				if err != nil {
					trashURLS.Set(url, currentNumberOfTries)
					todoURLS.Delete(url)
					fmt.Println(err)
					return
				} else {
					defer resp.Body.Close()
					if resp.StatusCode == 200 {
						numberOfURLSParsed++

						// Special case, download things in todo
						if download {
							contentType := resp.Header.Get("Content-type")
							contentTypes, contentTypeErr := mime.ExtensionsByType(contentType)
							extension := ""
							if contentTypeErr == nil {
								extension = contentTypes[0]
							}
							file_content, err := ioutil.ReadAll(resp.Body)
							if err != nil {
								log.Fatal(err)
							}

							var buf bytes.Buffer
							writer := gzip.NewWriter(&buf)
							writer.Write(file_content)
							writer.Close()
							filename := encodeURL(url) + extension + ".gz"

							err = ioutil.WriteFile(filename, buf.Bytes(), 0755)
							if err != nil {
								log.Fatal(err)
							}

							if logging {
								log.Printf("Saved %s to %s", url, encodeURL(url)+extension)
							}
							todoURLS.Delete(url)
							doneURLS.Set(url, currentNumberOfTries)
							return
						}

						links := collectlinks.All(resp.Body)
						if logging {
							log.Printf("Got %d links from %s\n", len(links), url)
						}
						for _, link := range links {
							if strings.Contains(link, "?") {
								link = strings.Split(link, "?")[0]
							}
							if !strings.Contains(link, "http") {
								link = baseURL + link
							}
							if !strings.Contains(link, baseURL) {
								continue
							}
							newLink, _ := urlx.Parse(link)
							urlNormalized, _ := urlx.Normalize(newLink)
							if len(urlNormalized) == 0 {
								continue
							}

							// Ignore keywords
							foundIgnoredKeyword := false
							for _, keyword := range keywordsToIgnore {
								if strings.Contains(urlNormalized, keyword) {
									foundIgnoredKeyword = true
									break
								}
							}
							if foundIgnoredKeyword {
								continue
							}

							// Include keywords
							foundIncludeKeyword := false
							for _, keyword := range keywordsToInclude {
								if strings.Contains(urlNormalized, keyword) {
									foundIncludeKeyword = true
									break
								}
							}
							if !foundIncludeKeyword {
								continue
							}

							if todoURLS.Get(urlNormalized, &foo) != nil &&
								doneURLS.Get(urlNormalized, &foo) != nil &&
								trashURLS.Get(urlNormalized, &foo) != nil {
								todoURLS.Set(urlNormalized, 0)
							}
						}
						doneURLS.Set(url, currentNumberOfTries)
						todoURLS.Delete(url)
					} else {
						if currentNumberOfTries > 3 {
							trashURLS.Set(url, currentNumberOfTries)
							todoURLS.Delete(url)
						} else {
							todoURLS.Set(url, currentNumberOfTries)
						}
					}
				}
			}(URLToDo)
		}
		wg.Wait()

		// Save every 5th iteration
		if math.Mod(float64(it), float64(5)) == 0 {
			if err = jsonstore.Save(doneURLS, "doneURLS.json"); err != nil {
				panic(err)
			}
			if err = jsonstore.Save(todoURLS, "todoURLS.json"); err != nil {
				panic(err)
			}
			if err = jsonstore.Save(trashURLS, "trashURLS.json"); err != nil {
				panic(err)
			}
			parsingTime := time.Since(programTime)
			timePerURL := parsingTime.Seconds() / float64(numberOfURLSParsed)
			numTodo := len(todoURLS.GetAll(regexp.MustCompile(`.*`)))
			numFinished := len(doneURLS.GetAll(regexp.MustCompile(`.*`)))
			log.Printf("Parsed %d urls in %s (%2.5f seconds / URL), %d/%d\n", numberOfURLSParsed, parsingTime.String(), timePerURL, numFinished, numTodo)
			if numTodo == 0 {
				os.Exit(0)
			}
		}
	}
}

func main() {
	// Parse flags
	var baseURL, ignoreKeywords, includeKeywords string
	var maxNumberWorkers, connectionPool int
	var logging, download bool
	flag.StringVar(&baseURL, "url", "", "URL to crawl `e.g. https://xkcd.com`")
	flag.StringVar(&ignoreKeywords, "ignore", "", "comma-delimited keywords to ignore")
	flag.StringVar(&includeKeywords, "include", "", "comma-delimited keywords that must include")
	flag.IntVar(&maxNumberWorkers, "workers", 100, "max number of workers")
	flag.IntVar(&connectionPool, "pool", 100, "number of connections in pool")
	flag.BoolVar(&logging, "v", false, "verbose")
	flag.BoolVar(&download, "dl", false, "download")
	flag.Parse()

	// Check it base URL is given
	if baseURL == "" && !download {
		print("Must specify URL to crawl")
		os.Exit(1)
	}

	// Determine the keywords to ignore
	var keywordsToIgnore []string
	if ignoreKeywords != "" {
		keywordsToIgnore = strings.Split(ignoreKeywords, ",")
		for i, keyword := range keywordsToIgnore {
			keywordsToIgnore[i] = strings.ToLower(strings.TrimSpace(keyword))
		}
	}

	// Determine the keywords to include
	var keywordsToInclude []string
	if includeKeywords != "" {
		keywordsToInclude = strings.Split(includeKeywords, ",")
		for i, keyword := range keywordsToInclude {
			keywordsToInclude[i] = strings.ToLower(strings.TrimSpace(keyword))
		}
	}

	// crawl
	crawl(baseURL, keywordsToIgnore, keywordsToInclude, maxNumberWorkers, connectionPool, logging, download)
}
