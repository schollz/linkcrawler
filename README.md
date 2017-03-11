# linkcrawler

![Coverage](https://img.shields.io/badge/coverage-76%25-green.svg?style=flat-square)
[![Doc](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/schollz/linkcrawler/lib)

Persistent and threaded web crawler that can either

  1. crawl a website and create a list of all links OR
  2. download all websites in a list to a gzipped file.

*linkcrawler* is threaded and uses connection pools so it is fast. It is persistent because it periodically dumps its state to JSON files which it will use to re-initialize if interrupted.

# Install

```
go get github.com/schollz/linkcrawler/...
```

# Run

## Crawl

To capture all the links on a website:

```bash
$ linkcrawler crawl http://rpiai.com
http://rpiai.com
2017/03/11 08:38:02 32 parsed (5/s), 0 todo, 32 done, 3 trashed
Got links downloaded from 'http://rpiai.com'
Wrote 32 links to NB2HI4B2F4XXE4DJMFUS4Y3PNU======.db.txt
```

## Download

To download gzipped webpages from a list of websites:

```bash
$ linkcrawler download *.txt
2017/03/11 08:41:20 32 parsed (31/s), 0 todo, 32 done, 0 trashed
Finished downloading
$ ls downloaded | head -n 2
NB2HI4B2F4XXE4DJMFUS4Y3PNU======.html.gz
NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====.html.gz
```

Downloads are saved into a folder `downloaded` with url of link encoded in Base32.

## Persistence

The current state of the crawler is saved into a BoltDB database and also backed up in to a Zip-archive. If the crawler is interrupted, you can simply run the command again and it will restart from the last state. To dump the current database, just use

```bash
$ linkcrawler.exe dump NB2HI4B2F4XXE4DJMFUS4Y3PNU======.db
Got links downloaded from 'http://rpiai.com'
Wrote 32 links to NB2HI4B2F4XXE4DJMFUS4Y3PNU======.db.txt
```


# License

MIT
