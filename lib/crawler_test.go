package crawler

import (
	"testing"

	"github.com/schollz/boltdb-server/connect"
)

func TestGeneral(t *testing.T) {
	boltdbserver := "http://localhost:8080"

	// Delete previous
	conn, _ := connect.Open(boltdbserver, "linkcrawler")
	_ = conn.DeleteDatabase()

	crawl, err := New("http://rpiai.com/", boltdbserver)
	if err != nil {
		t.Error(err)
	}
	crawl.Verbose = true

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
	crawl, err = New("http://rpiai.com/", boltdbserver)
	if err != nil {
		t.Error(err)
	}
	crawl.Verbose = true

	err = crawl.Download(allLinks)
	if err != nil {
		t.Errorf("Problem downloading: %s", err.Error())
	}

	err = crawl.Dump()
	if err != nil {
		t.Errorf("Problem dumping: %s", err.Error())
	}
}
