
<p align="center">
<img
    src="logo.png"
    width="260" height="80" border="0" alt="linkcrawler">
<br>
<a href="https://travis-ci.org/schollz/linkcrawler"><img src="https://img.shields.io/travis/schollz/linkcrawler.svg?style=flat-square" alt="Build Status"></a>
<a href="http://gocover.io/github.com/schollz/linkcrawler/lib"><img src="https://img.shields.io/badge/coverage-76%25-yellow.svg?style=flat-square" alt="Code Coverage"></a>
<a href="https://godoc.org/github.com/schollz/linkcrawler/lib"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>
</p>

<p align="center">Cross-platform persistent and distributed web crawler</a></p>

*linkcrawler* is persistent because the queue is stored in a remote database that is automatically re-initialized if interrupted. *linkcrawler* is distributed because multiple instances of *linkcrawler* will work on the remotely stored queue, so you can start as many crawlers as you want on separate machines to speed along the process. *linkcrawler* is also fast because it is threaded and uses connection pools.

Crawl responsibly.

# This repo has been superseded by [schollz/goredis-crawler](https://github.com/schollz/goredis-crawler)

Getting Started
===============

## Install

If you have Go installed, just do
```
go get github.com/schollz/linkcrawler/...
go get github.com/schollz/boltdb-server/...
```

Otherwise, use the releases and [download linkcrawler](https://github.com/schollz/linkcrawler/releases/latest) and then [download the boltdb-server](https://github.com/schollz/boltdb-server/releases/latest).


## Run

### Crawl a site

First run the database server which will create a LAN hub:

```sh
$ ./boltdb-server
boltdb-server running on http://X.Y.Z.W:8050
```

Then, to capture all the links on a website:

```sh
$ linkcrawler --server http://X.Y.Z.W:8050 crawl http://rpiai.com
```


Make sure to replace `http://X.Y.Z.W:8050` with the IP information outputted from the boltdb-server.

You can run this last command on as many different machines as you want, which will help to crawl the respective website and add collected links to a universal queue in the server.

The current state of the crawler is saved. If the crawler is interrupted, you can simply run the command again and it will restart from the last state.

See the help (`-help`) if you'd like to see more options, such as exclusions/inclusions and modifying the worker pool and connection pools.


### Download a site

You can also use *linkcrawler* to download webpages from a newline-delimited list of websites. As before, first startup a boltdb-server.  Then you can run:

```bash
$ linkcrawler --server http://X.Y.Z.W:8050 download links.txt
```

Downloads are saved into a folder `downloaded` with URL of link encoded in Base32 and compressed using gzip.

### Dump the current list of links

To dump the current database, just use

```bash
$ linkcrawler --server http://X.Y.Z.W:8050 dump http://rpiai.com
Wrote 32 links to NB2HI4B2F4XXE4DJMFUS4Y3PNU======.txt
```

## License

MIT
