# sesopenko/s3sync

A simple S3 synchronization app which also syncs object metadata. I couldn't find an s3
synchronization solution pre-built which copies metadata along with the files, so I built one. I've only put about an
afternoon into this project so your own mileage may vary.

## What It Does

When the program runs it will synchronize any files from the remote bucket to the local machine. It skips files older
than 1 year. It will then sleep a minute and redo the sync. Object metadata will be stored as `<filename>.details.txt`.

Object keys and filenames are cleaned for compatibility with windows file paths. Invalid characters, such as `*`, will
be converted to `_`.

## Cost Estimates

It's estimated that the per-minute S3 list objects operation  will cost 24 cents per month for every 1000 items in the
bucket. It's a brute force approach compared to listening  for events on the bucket, but it's far simpler to set up.

## Developer Notes

See [DEVNOTES.md](DEVNOTES.md) for environment variables which need to be set in order for the app to run.

## Required Runtime Environment variables

* AWS_ACCESS_KEY_ID=<aws access key>
* AWS_SECRET_ACCESS_KEY=<aws secret key>
* AWS_REGION=<default region>
* BUCKET=<s3 bucket name>
* MAIN_FOLDER=<folder to synchronize>
* SAVE_PATH=/mnt/dest

## Docker

[https://hub.docker.com/r/sesopenko/s3sync](https://hub.docker.com/r/sesopenko/s3sync)

### Docker Run Example

This example will synchronize the files and store them in the host computer's `/mnt/dest` path.

```bash
docker run -it sesopenko/s3sync \
-v ./mnt/dest:/mnt/dest \
-e AWS_ACCESS_KEY_ID=YOURKEY \
-e AWS_REGION=us-west-2 \
-e AWS_SECRET_ACCESS_KEY=YOURSECRETKEY \
-e BUCKET=your-bucket-name-do-not-use-arn
-e MAIN_FOLDER=thesub/folder/your/files/are/in/can/be/blank
-e SAVE_PATH=/mnt/dest
```

## Licensed GNU GPL Version 3

This software is licensed GNU GPL Version 3 which may be read in [LICENSE.md](LICENSE.md) or [LICENSE.txt](LICENSE.txt).
A copy of this license should be included in this projectd. If one is not included it may be downloaded
from [https://www.gnu.org/licenses/gpl-3.0.txt](https://www.gnu.org/licenses/gpl-3.0.txt)

## Copyright

Copyright © Sean Esopenko 2023 All Rights Reserved