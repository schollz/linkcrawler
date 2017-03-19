
<p align="center">
<img
    src="logo.png"
    width="260" height="80" border="0" alt="linkcrawler">
<br>
<a href="https://travis-ci.org/schollz/linkcrawler"><img src="https://img.shields.io/travis/schollz/linkcrawler.svg?style=flat-square" alt="Build Status"></a>
<a href="http://gocover.io/github.com/schollz/linkcrawler/lib"><img src="https://img.shields.io/badge/coverage-76%25-yellow.svg?style=flat-square" alt="Code Coverage"></a>
<a href="https://godoc.org/github.com/schollz/linkcrawler/lib"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>
</p>

<p align="center">Persistent and distributed web crawler</a></p>

Persistent and distributed web crawler that can either crawl a website and create a list of all links OR download all websites in a list to a gzipped file. *linkcrawler* is threaded and uses connection pools so it is fast. It is persistent because it periodically dumps its state to JSON files which it will use to re-initialize if interrupted. It is distributed by connecting to a database to store its state so you can start as many crawlers as you want on separate machines to speed along the process.

Crawl responsibly.

Getting Started
===============

## Install

If you have Go installed, just do
```
go get github.com/schollz/linkcrawler/...
go get github.com/schollz/boltdb-server/...
```

Otherwise, use the releases and [download linkcrawler](https://github.com/schollz/linkcrawler/releases/latest) and then [download boltdb-server](https://github.com/schollz/boltdb-server/releases/latest)


## Run

### Crawl

First run the database server which will create a hub:

```sh
$ ./boltdb-server
boltdb-server running on X.Y.Z.W:8080
```

Then, to capture all the links on a website:

```sh
$ linkcrawler --server http://X.Y.Z.W:8080 crawl http://rpiai.com
http://rpiai.com
Setting up crawler...
Starting crawl using DB NB2HI4B2F4XXE4DJMFUS4Y3PNU======
2017/03/11 08:38:02 32 parsed (5/s), 0 todo, 32 done, 3 trashed
Got links downloaded from 'http://rpiai.com'
Wrote 32 links to NB2HI4B2F4XXE4DJMFUS4Y3PNU======.txt
```


Make sure to replace `http://X.Y.Z.W:8080` with the IP information outputed from the boltdb-server.

You can run this last command on as many different machines as you want, which will help to crawl the respective website and add collected links to a universal queue in the server.

The current state of the crawler is saved. If the crawler is interrupted, you can simply run the command again and it will restart from the last state.


### Download

Similar as before, startup a boltdb-server and then, to download gzipped webpages from a list of websites:

```bash
$ linkcrawler --server http://X.Y.Z.W:8080 download links.txt
2017/03/11 08:41:20 32 parsed (31/s), 0 todo, 32 done, 0 trashed
Finished downloading
$ ls downloaded | head -n 2
NB2HI4B2F4XXE4DJMFUS4Y3PNU======.html.gz
NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====.html.gz
```

Downloads are saved into a folder `downloaded` with URL of link encoded in Base32.

### Dump

To dump the current database, just use

```bash
$ linkcrawler --server http://X.Y.Z.W:8080 dump http://rpiai.com
Wrote 32 links to NB2HI4B2F4XXE4DJMFUS4Y3PNU======.txt
```


## License

MIT
