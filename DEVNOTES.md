# DEVNOTES

## Useful links

* [S3 headobject](https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#S3.HeadObject)
** Used to retrieve object metadata

## How to remove funky windows dirs

1. cd to directory
2. `dir /x` to get useable labels
3. for each: `rmdir /q /s SOMEDIR~1`