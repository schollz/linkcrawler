package crawler

import (
	"os"
	"regexp"
	"testing"
)

func TestGeneral(t *testing.T) {
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====_crawl_todo.json")
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====_crawl_done.json")
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====_crawl_trash.json")
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====_dl_todo.json")
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====_dl_done.json")
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====_dl_trash.json")
	crawl, err := New("http://rpiai.com/")
	if err != nil {
		t.Error(err)
	}
	crawl.Verbose = true

	if err = crawl.Crawl(); err != nil {
		t.Error(err)
	}

	numTodo, err := crawl.saveKeyStores(false)
	if err != nil {
		t.Error(err)
	}
	if numTodo != 0 {
		t.Errorf("numTodo should be 0 but it is %d", numTodo)
	}

	if len(crawl.done.GetAll(regexp.MustCompile(`.*`))) < 30 {
		t.Errorf("Did not get all the websites crawled!")
	}

	allLinks := crawl.GetLinks()
	if len(allLinks) < 30 {
		t.Errorf("Only got %d links", len(allLinks))
	}

	err = crawl.Download(allLinks)
	if err != nil {
		t.Errorf("Problem downloading: %s", err.Error())
	}
}
