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

### bundles3-index

Generate and store a plain-text index of all the bundles (latest and dated) in an S3 bucket.

```
$> bundles3-index whosonfirst.mapzen.com us-east-1
upload: ../../../../tmp/bundles-index.txt to s3://whosonfirst.mapzen.com/bundles/index.txt
```

## See also

* https://aws.amazon.com/cli/