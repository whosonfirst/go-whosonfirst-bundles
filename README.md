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
#!/bin/sh

BUNDLE_TOOL="/usr/local/bin/go-whosonfirst-bundles/bin/wof-bundle-metafiles"
PRUNE_TOOL="/usr/local/bin/go-whosonfirst-bundles/bin/wof-prune-bundles"
INDEX_TOOL="/usr/local/bin/go-whosonfirst-bundles/utils/bundles3-index"

BUCKET="whosonfirst.mapzen.com"
PREFIX="bundles"
REGION="us-east-1"

BUNDLES="/usr/local/data/bundles"

# by default limits each metafile to 10 bundles (wof-county-20170702, wof-county-20170703 and so on)

${PRUNE_TOOL} -root ${BUNDLES}
${PRUNE_TOOL} -remote -bucket ${BUCKET} -prefix ${PREFIX} -region ${REGION}

for REPO in $@
do
    ${BUNDLE_TOOL} -dated -latest -compress -dest ${BUNDLES} ${REPO}
done

aws s3 sync --region ${REGION} --exclude "*" --include "*.bz2" --include "*.sha1.txt" ${BUNDLES} s3://${BUCKET}/${PREFIX}/

${INDEX_TOOL} ${BUCKET} ${REGION}
exit 0
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
