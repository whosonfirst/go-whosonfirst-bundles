# utils

## bundles3-*

Tools for cloning, pruning and indexing bundles in S3.

### bundles3-prune-dated

Prune any dated bundles or metafiles (see below) older than 14 days.

```
$> bundles3-prune-dated whosonfirst.mapzen.com us-east-1
prune wof-continent-20170703T201836.csv.bz2
...and so on
```

### bundles3-latest-to-dated

Make "dated" copies of all the "latest" bundles and metafiles. For example `wof-continent-latest.csv.bz2` is copied to `wof-continent-20170728T201836.csv.bz2` (because the former's creation date is `2017-07-28 20:18:36`).

```
$> bundles3-latest-to-dated whosonfirst.mapzen.com us-east-1
copy: s3://whosonfirst.mapzen.com/bundles/wof-continent-latest.csv.bz2 to s3://whosonfirst.mapzen.com/bundles/wof-continent-20170728T201836.csv.bz2
copy: s3://whosonfirst.mapzen.com/bundles/wof-continent-latest.csv.bz2.sha1.txt to s3://whosonfirst.mapzen.com/bundles/wof-continent-20170728T201836.csv.bz2.sha1.txt
...and so on
```

### bundles3-index

Generate and store a plain-text index of all the bundles (latest and dated) in an S3 bucket.

```
$> bundles3-index whosonfirst.mapzen.com us-east-1
upload: ../../../../tmp/bundles-index.txt to s3://whosonfirst.mapzen.com/bundles/index.txt
```

## See also

* https://aws.amazon.com/cli/