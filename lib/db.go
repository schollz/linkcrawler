package crawler

import (
	"encoding/base32"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
)

// GetLinks a list of the done urls
func (c *Crawler) getNLinksTodo(n int) ([]string, error) {
	links := make([]string, n)

	db, err := bolt.Open(c.FilePrefix+".db", 0600, nil)
	if err != nil {
		return links, err
	}
	defer db.Close()

	i := 0
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("todo"))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			links[i] = string(k)
			i++
			if i == n {
				break
			}
		}
		return nil
	})
	if i < n {
		links = links[0:i]
	}
	return links, nil
}

func Dump(db string) error {
	dbnameBase32 := strings.Split(db, ".")[0]
	originalLink, err := base32.StdEncoding.DecodeString(dbnameBase32)
	if err != nil {
		return err
	}
	fmt.Println(string(originalLink))
	crawl, err := New("http://rpiai.com/")
	if err != nil {
		return err
	}

	links, err := crawl.GetLinks()
	if err != nil {
		return err
	}
	fmt.Println(strings.Join(links, "\n"))

	return nil
}

// GetLinks a list of the done urls
func (c *Crawler) GetLinks() ([]string, error) {
	var links []string
	db, err := bolt.Open(c.FilePrefix+".db", 0600, nil)
	if err != nil {
		return links, err
	}
	defer db.Close()

	numberOfLinks := 0
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("todo"))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			numberOfLinks++
		}

		b = tx.Bucket([]byte("done"))
		c = b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			numberOfLinks++
		}
		return nil
	})

	links = make([]string, numberOfLinks)
	numberOfLinks = 0
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("todo"))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			links[numberOfLinks] = string(k)
			numberOfLinks++
		}

		b = tx.Bucket([]byte("done"))
		c = b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			links[numberOfLinks] = string(k)
			numberOfLinks++
		}
		return nil
	})

	return links, nil
}

// getAllKeys a list of the done urls
func (c *Crawler) getAllKeys(bucket string) ([]string, error) {
	var links []string
	db, err := bolt.Open(c.FilePrefix+".db", 0600, nil)
	if err != nil {
		return links, err
	}
	defer db.Close()

	numberOfLinks := 0
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			numberOfLinks++
		}
		return nil
	})

	links = make([]string, numberOfLinks)
	numberOfLinks = 0
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			links[numberOfLinks] = string(k)
			numberOfLinks++
		}
		return nil
	})

	return links, nil
}

func (c *Crawler) initDB() error {
	// Return if already exists
	if _, err := os.Stat(c.FilePrefix + ".db"); err == nil {
		links, _ := c.getAllKeys("todo")
		c.numToDo = len(links)
		links, _ = c.getAllKeys("done")
		c.numDone = len(links)
		links, _ = c.getAllKeys("trash")
		c.numTrash = len(links)
		return nil
	}

	// Create a new DB
	db, err := bolt.Open(c.FilePrefix+".db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("todo"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucket([]byte("done"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucket([]byte("trash"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func (c *Crawler) set(bucket string, key string, numTries int) error {
	db, err := bolt.Open(c.FilePrefix+".db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		numTriesString := strconv.Itoa(numTries)
		err := b.Put([]byte(key), []byte(numTriesString))
		return err
	})
}

func (c *Crawler) contains(bucket string, key string) (bool, error) {
	db, err := bolt.Open(c.FilePrefix+".db", 0600, nil)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var numTriesByte []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		numTriesByte = b.Get([]byte(key))
		return nil
	})

	if numTriesByte != nil {
		return true, nil
	}
	return false, nil
}

func (c *Crawler) get(bucket string, key string) (int, error) {
	db, err := bolt.Open(c.FilePrefix+".db", 0600, nil)
	if err != nil {
		return -1, err
	}
	defer db.Close()

	numTries := 0
	var numTriesByte []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		numTriesByte = b.Get([]byte(key))
		return nil
	})

	if numTriesByte != nil {
		numTries, _ = strconv.Atoi(string(numTriesByte))
	}
	return numTries, nil
}

func (c *Crawler) delete(bucket string, key string) error {
	db, err := bolt.Open(c.FilePrefix+".db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Delete([]byte(key))
	})
}
