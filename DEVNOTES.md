# DEVNOTES

## Runtime Environment variables:

* AWS_ACCESS_KEY_ID=<aws access key>
* AWS_SECRET_ACCESS_KEY=<aws secret key>
* AWS_REGION=<default region>
* BUCKET=<s3 bucket name>
* MAIN_FOLDER=<folder to synchronize>
* SAVE_PATH=/mnt/dest

## Docker mounts

Mount the container path the $SAVE_PATH environment path to a local path to access the files.

## Useful links

* [S3 headobject](https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#S3.HeadObject)
** Used to retrieve object metadata