
<p align="center">
<img 
    src="logo.png" 
    width="240" height="78" border="0" alt="linkcrawler">
<br>
<a href="https://travis-ci.org/schollz/linkcrawler"><img src="https://img.shields.io/travis/schollz/linkcrawler.svg?style=flat-square" alt="Build Status"></a>
<a href="http://gocover.io/github.com/schollz/linkcrawler/lib"><img src="https://img.shields.io/badge/coverage-76%25-yellow.svg?style=flat-square" alt="Code Coverage"></a>
<a href="https://godoc.org/github.com/schollz/linkcrawler/lib"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>
</p>

<p align="center">Persistent and threaded web crawler</a></p>

Persistent and threaded web crawler that can either

  1. crawl a website and create a list of all links OR
  2. download all websites in a list to a gzipped file.

*linkcrawler* is threaded and uses connection pools so it is fast. It is persistent because it periodically dumps its state to JSON files which it will use to re-initialize if interrupted.

Getting Started
===============

## Install

```
go get github.com/schollz/linkcrawler/...
```

## Run

### Crawl

To capture all the links on a website:

```bash
$ linkcrawler crawl http://rpiai.com
http://rpiai.com
2017/03/11 08:38:02 32 parsed (5/s), 0 todo, 32 done, 3 trashed
Got links downloaded from 'http://rpiai.com'
Wrote 32 links to NB2HI4B2F4XXE4DJMFUS4Y3PNU======.db.txt
```

### Download

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

### Persistence

The current state of the crawler is saved into a local database and also backed up into a Zip-archive periodically. If the crawler is interrupted, you can simply run the command again and it will restart from the last state:

```bash
$ linkcrawler.exe crawl http://rpiai.com
http://rpiai.com
2017/03/11 08:47:13 18 parsed (6/s), 23 todo, 17 done, 0 trashed
Ctl+C
$ linkcrawler.exe crawl http://rpiai.com  # It will read the previous DB
http://rpiai.com
2017/03/11 08:47:18 13 parsed (13/s), 15 todo, 19 done, 5 trashed # Continued!
Got links downloaded from 'http://rpiai.com'
Wrote 32 links to NB2HI4B2F4XXE4DJMFUS4Y3PNU======.db.txt
```

### Dump

To dump the current database, just use

```bash
$ linkcrawler.exe dump NB2HI4B2F4XXE4DJMFUS4Y3PNU======.db
Got links downloaded from 'http://rpiai.com'
Wrote 32 links to NB2HI4B2F4XXE4DJMFUS4Y3PNU======.db.txt
```


## License

MIT
