# linkcrawler

![Coverage](https://img.shields.io/badge/coverage-68%25-green.svg?style=flat-square)
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

To capture all the links on a website:

```
$ linkcrawler crawl http://rpiai.com
http://rpiai.com
2017/03/09 21:37:35 Parsed 32 urls in 2.8017983s (11 URLs/s). Finished: 32, Todo: 0
32 links written to links.txt
```

To download gzipped webpages:
```
$ linkcrawler download links.txt
2017/03/09 21:38:22 Parsed 32 urls in 655.7632ms (48 URLs/s). Finished: 32, Todo: 0
Finished downloading$ ls downloaded | head -n 2
NB2HI4B2F4XXE4DJMFUS4Y3PNU======.html.gz
NB2HI4B2F4XXE4DJMFUS4Y3PNUXQ====.html.gz
```

Downloads are saved into a folder `downloaded` with url of link encoded in Base32.


## Persistence

The current state of the crawler is saved into three JSON files (`XYZ_crawl_done|todo|trash.json`, where `XYZ` is the link/file encoded as Base32). If the crawler is interrupted, you can simply run the command again and it will restart using the respective files as the state. You can also remove these files to have it start from scratch. The amount of persistence can be controlled using `-save`:

```
$ linkcrawler -save 1 crawl http://rpiai.com
http://rpiai.com
2017/03/09 08:00:15 Parsed 1 urls in 385.4047ms (0.38540 seconds / URL), 1/17
2017/03/09 08:00:16 Parsed 18 urls in 1.0933271s (0.06074 seconds / URL), 18/21
Ctl+C
$ linkcrawler -save 1 crawl http://rpiai.com
http://rpiai.com
2017/03/09 08:00:19 Parsed 14 urls in 321.9438ms (0.02300 seconds / URL), 32/3
2017/03/09 08:00:19 Parsed 14 urls in 510.4331ms (0.03646 seconds / URL), 32/3
2017/03/09 08:00:20 Parsed 14 urls in 714.087ms (0.05101 seconds / URL), 32/3
2017/03/09 08:00:20 Parsed 14 urls in 904.1514ms (0.06458 seconds / URL), 32/1
2017/03/09 08:00:20 Parsed 14 urls in 1.0947706s (0.07820 seconds / URL), 32/0
32 links written to links.txt
```

# License

MIT
