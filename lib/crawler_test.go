package crawler

import (
	"os"
	"testing"
)

func TestGeneral(t *testing.T) {
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====.db")
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====.db.zip")
	crawl, err := New("http://rpiai.com/")
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

	err = crawl.Download(allLinks)
	if err != nil {
		t.Errorf("Problem downloading: %s", err.Error())
	}
}
