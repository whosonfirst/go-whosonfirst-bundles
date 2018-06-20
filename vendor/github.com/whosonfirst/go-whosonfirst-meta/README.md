# go-whosonfirst-meta

Go package for working with Who's On First meta files

## Install

You will need to have both `Go` and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Tools

### wof-build-metafiles

Crawl a Who's On First data repository and rebuild one or more meta files.

```
$> wof-build-metafiles -h
Usage of ./bin/wof-build-metafiles:
  -exclude string
    	A comma-separated list of placetypes that meta files will not be created for.
  -open-filehandles int
    	The maximum number of file handles to keep open at any given moment. (default 512)
  -out string
    	Where to store metafiles. If empty then assume metafile are created in a child folder of 'repo' called 'meta'.
  -placetypes string
    	A comma-separated list of placetypes that meta files will be created for. All other placetypes will be ignored.
  -processes int
    	The number of concurrent processes to use. (default 8)
  -repo string
    	Where to read data (to create metafiles) from. If empty then the code will assume the current working directory.
  -roles string
    	Role-based filters are not supported yet.
```

For example:

```
$> cd whosonfirst-data
$> wof-build-metafiles
2017/06/29 14:28:51 time to dump 486685 features: 3m46.345202842s

```

## See also

* https://github.com/whosonfirst/py-mapzen-whosonfirst-meta
