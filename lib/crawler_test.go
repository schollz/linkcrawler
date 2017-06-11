package crawler

import (
	"fmt"
	"testing"

	"github.com/schollz/boltdb-server/connect"
)

func TestGeneral(t *testing.T) {
	boltdbserver := "http://localhost:8050"
	crawl, err := New("http://rpiai.com/", boltdbserver, true)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(crawl.BaseURL)
	conn, _ := connect.Open(boltdbserver, crawl.Name())

	// Delete previous
	_ = conn.DeleteDatabase()
	if err != nil {
		t.Error(err)
	}

	crawl, err = New("http://rpiai.com/", boltdbserver, true)
	if err != nil {
		t.Error(err)
	}

	if err = crawl.Crawl(); err != nil {
		t.Error(err)
	}

	allLinks, err := crawl.GetLinks()
	if err != nil {
		t.Error(err)
	}
	if len(allLinks) < 30 {
		t.Errorf("Only got %d links", len(allLinks))
	}

	// Reload the crawler
	conn.DeleteDatabase()
	crawl, err = New("http://rpiai.com/", boltdbserver, true)
	if err != nil {
		t.Error(err)
	}

	err = crawl.Download(allLinks)
	if err != nil {
		t.Errorf("Problem downloading: %s", err.Error())
	}

	err = crawl.Dump()
	if err != nil {
		t.Errorf("Problem dumping: %s", err.Error())
	}
}
