# linkcrawler

Persistent and threaded web crawler that can either A) download a list of all links on a website or B) download a list of websites. It is threaded and uses connection pools so it is fast. It is persistent because it periodically dumps its state to JSON files which can be restarted from at any time.


```
$ go build 
$ ./gogetlinks -url 'https://xkcd.com'
$ wc -l *json
 1820 doneURLS.json
    1 todoURLS.json
    4 trashURLS.json
```

Another usage, to download concurrently:

```
$ cat todoURLS.json
{
    "https://www.example.org":"0"
}
$ gogetlinks -dl  # Download URL keys in todoURLS.json, 
                  # skips if they are already downloaded.
                  # Periodically writes to doneURLS.json so it can be 
                  # restarted.
$ ls | grep gz    # Files saved as gzipped, with extensions inferred 
f486ce7caeffaffbdbd6a028fa561f1af5fe3380157f8263e71f1edb7ff06f3f.html.gz
``` 
