# go-whosonfirst-repo

Go package for working with Who's On First repos

## Install

You will need to have both `Go` and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Tools

### wof-repo-parse

```
./bin/wof-repo-parse -h
Usage of ./bin/wof-repo-parse:
  -dated
    	Use the current YYYYMMDD date as the suffix for a filename
  -name value
    	One or more canned filename types. Valid options are: bundle, concordances, meta, sqlite or all
  -old-skool
    	Use old skool 'wof-' prefix when generating filenames
  -placetype string
    	Specify a specific placetype when generating a filename
```

For example:

```
$> ./bin/wof-repo-parse -name all -placetype region whosonfirst-data
whosonfirst-data    OK
sqlite filename	    whosonfirst-data-region-latest.db
meta filename	    whosonfirst-data-region-latest.csv
bundle filename	    whosonfirst-data-region-latest
concordances filename	whosonfirst-data-region-concordances.csv

$> ./bin/wof-repo-parse -name all whosonfirst-data-venue-ca
whosonfirst-data-venue-ca  OK
sqlite filename		   whosonfirst-data-venue-ca-latest.db
meta filename		   whosonfirst-data-venue-ca-latest.csv
bundle filename		   whosonfirst-data-venue-ca-latest
concordances filename	   whosonfirst-data-venue-ca-concordances.csv

$> ./bin/wof-repo-parse -name all -placetype region -dated whosonfirst-data
whosonfirst-data     OK
sqlite filename	     whosonfirst-data-region-20180724.db
meta filename	     whosonfirst-data-region-20180724.csv
bundle filename	     whosonfirst-data-region-20180724
concordances filename	whosonfirst-data-region-concordances.csv

$> ./bin/wof-repo-parse -name all -placetype region -old-skool whosonfirst-data
whosonfirst-data     OK
sqlite filename	     wof-region-latest.db
meta filename	     wof-region-latest.csv
bundle filename	     wof-region-latest
concordances filename	wof-region-concordances.csv
```