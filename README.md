# go-whosonfirst-bundles

Go package for working with Who's On First data bundles

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Tools

For example:

```
./bin/wof-prune-bundles -remote -bucket whosonfirst.mapzen.com -prefix bundles -max-bundles 10
./bin/wof-prune-bundles -root /usr/local/data/bundles -max-bundles 10
./bin/wof-bundle-metafiles -dated -latest -compress -dest /usr/local/data/bundles /usr/local/data/whosonfirst-data
aws s3 sync --region us-east-1 /usr/local/data/bundles s3://whosonfirst.mapzen.com/bundles/
```

### wof-bundle-metafiles

_Please write me._

```
./bin/wof-bundle-metafiles -source file:///usr/local/data/whosonfirst-data-venue-lagov/data -name 3co-venue-lagov -dated -dest ./ -compress /usr/local/data/whosonfirst-data-venue-lagov/meta/wof-whosonfirst-data-venue-lagov-latest.csv
2017/07/18 19:22:19 /usr/local/mapzen/go-whosonfirst-bundles/3co-venue-lagov-20170718.tar.bz2
```

### wof-prune-bundles

_Please write me._

## See also

* https://github.com/whosonfirst/go-whosonfirst-clone
