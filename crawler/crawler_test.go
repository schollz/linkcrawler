package crawler

import (
	"os"
	"regexp"
	"testing"
)

func TestGeneral(t *testing.T) {
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====_todo.json")
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====_done.json")
	defer os.Remove("NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====_trash.json")
	crawl, err := New("http://rpiai.com/")
	if err != nil {
		t.Error(err)
	}
	crawl.Verbose = true

	if err = crawl.Crawl(); err != nil {
		t.Error(err)
	}

	numTodo, err := crawl.saveKeyStores()
	if err != nil {
		t.Error(err)
	}
	if numTodo != 0 {
		t.Errorf("numTodo should be 0 but it is %d", numTodo)
	}

	if len(crawl.done.GetAll(regexp.MustCompile(`.*`))) < 30 {
		t.Errorf("Did not get all the websites crawled!")
	}
}
