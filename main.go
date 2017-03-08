package main

import (
	"flag"
	"fmt"
	"log"
	"math"
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

func crawl(baseURL string, keywordsToIgnore []string, maxNumberWorkers int, connectionPool int, logging bool) {
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
	if _, err = os.Stat("doneURLS.json"); err == nil {
		var err2 error
		doneURLS, err2 = jsonstore.Open("doneURLS.json")
		if err2 != nil {
			panic(err)
		}
		todoURLS, err2 = jsonstore.Open("todoURLS.json")
		if err2 != nil {
			panic(err)
		}
		trashURLS, err2 = jsonstore.Open("trashURLS.json")
		if err2 != nil {
			panic(err)
		}
	} else {
		todoURLS.Set(baseURL, true)
	}

	// Start scraping
	programTime := time.Now()
	numberOfURLSParsed := 0
	it := 0
	for {
		it++
		numWorkers := 0
		for URLToDo := range todoURLS.GetAll(regexp.MustCompile(`.*`)) {
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

							if todoURLS.Get(urlNormalized, &foo) != nil &&
								doneURLS.Get(urlNormalized, &foo) != nil &&
								trashURLS.Get(urlNormalized, &foo) != nil {
								todoURLS.Set(urlNormalized, 0)
							}
						}
						doneURLS.Set(url, currentNumberOfTries)
						todoURLS.Delete(url)
						numberOfURLSParsed++
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
	var baseURL, ignoreKeywords string
	var maxNumberWorkers, connectionPool int
	var logging bool
	flag.StringVar(&baseURL, "url", "", "URL to crawl `e.g. https://xkcd.com`")
	flag.StringVar(&ignoreKeywords, "ignore", "", "comma-delimited keywords to ignore")
	flag.IntVar(&maxNumberWorkers, "workers", 100, "max number of workers")
	flag.IntVar(&connectionPool, "pool", 100, "number of connections in pool")
	flag.BoolVar(&logging, "v", false, "verbose")
	flag.Parse()

	// Check it base URL is given
	if baseURL == "" {
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

	// crawl
	crawl(baseURL, keywordsToIgnore, maxNumberWorkers, connectionPool, logging)
}
