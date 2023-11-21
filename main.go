package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type FakeWriterAt struct {
	w io.Writer
}

func (fw FakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we forced sequential downloads
	return fw.w.Write(p)
}

func main() {
	startTime := time.Now()
	restartTime := startTime.AddDate(0, 0, 7)
	for true {
		syncBucket()
		time.Sleep(time.Minute)
		if time.Now().After(restartTime) {
			exitErrorf("Restarting after running for 7 days to avoid memory leak errors.")
		}
	}
}

func syncBucket() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg)
	const maxKeys = 1000

	bucket := os.Getenv("BUCKET")
	listInput := s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(maxKeys),
		Prefix:  aws.String(os.Getenv("MAIN_FOLDER")),
	}

	walkBucketFiles(client, &listInput)
}

func handleFile(client *s3.Client, bucket string, key string, size int64) error {
	savePath := os.Getenv("SAVE_PATH")
	sanitizedKey := sanitizeWindowsPath(key)
	sizeMb := float64(size) / 1024.0 / 1024.0
	filePath := fmt.Sprintf("%s/%s", savePath, sanitizedKey)

	// if filePath doesn't exist
	// download to location
	if fi, fileErr := os.Stat(filePath); fileErr == nil {
		// path to filePath exists
		if fi.Size() == size {
			return nil
		} else {
			log.Println("File is incorrect size.  Deleting and downloading again.")
			e := os.Remove(filePath)
			if e != nil {
				return e
			}
		}
	} else if errors.Is(fileErr, os.ErrNotExist) {
		log.Printf("filePath does not exist: %s", filePath)
	} else {
		// some other error
		log.Printf("Unable to  check filePath exists: %s (%s)", filePath, fileErr)
		return fileErr
	}
	// path doesn't exist, let's create it
	dir := filepath.Dir(filePath)
	if dirErr := os.MkdirAll(dir, 0777); dirErr != nil {
		log.Printf("Unable to create directory: %s", dir)
		log.Print(dirErr)
	}
	log.Printf("downloading filePath: %s (%.2f mb)", filePath, sizeMb)
	result, downErr := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if downErr != nil {
		log.Printf("Couldn't get object %v:%v. Error: %v\n", bucket, key, downErr)
		return downErr
	}
	writeMeta(dir, result)

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Could not create file: %v", err)
		return err
	}
	body, bodyErr := io.ReadAll(result.Body)
	if bodyErr != nil {
		log.Printf("Couldn't read object body from %v. Error: %v\n", key, bodyErr)
		return bodyErr
	}
	_, err = file.Write(body)
	if err != nil {
		log.Printf("Couldn't write file body from %v. Error: %v\n", key, err)
	}
	file.Sync()

	defer file.Close()
	// if dir doesn't exist
	// create it

	return nil
}

func sanitizeWindowsPath(key string) string {
	var sanitizedKey = key
	// The following characters are invalid for Windows systems.
	// Not filtering forward slashes because this typically runs in a posix environment
	// which will interpret those as directory separators.
	sanitizedKey = strings.ReplaceAll(sanitizedKey, "?", "_")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, "<", "_")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, ">", "_")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, ":", "_")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, "\"", "_")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, "\\", "_")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, "|", "_")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, "?", "_")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, "*", "_")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, " ", "_")
	return sanitizedKey
}

func writeMeta(dir string, object *s3.GetObjectOutput) {
	metaPath := path.Join(dir, "details.txt")
	if _, fileErr := os.Stat(metaPath); fileErr == nil {
		// path to filePath exists
		log.Printf("meta filePath exists: %s", metaPath)
		return
	} else if errors.Is(fileErr, os.ErrNotExist) {
		log.Printf("writing meta")
		metaString := ""
		dec := new(mime.WordDecoder)
		for k, v := range object.Metadata {
			kt := strings.TrimSpace(k)
			// The values are encoded using http header encoding. Need to decode them so that
			// unicode characters (like emojis) come through.
			vt, err := dec.DecodeHeader(v)
			if err != nil {
				log.Printf("Unable to decode header (%v)", err)
				vt = v
			}
			vt = strings.TrimSpace(vt)
			metaString += fmt.Sprintln(fmt.Sprintf("%s: %s", kt, vt))
		}
		file, err := os.Create(metaPath)
		if err != nil {
			log.Printf("Could not create meta file: %v", err)
			return
		}
		file.WriteString(metaString)
		file.Sync()
		defer file.Close()
	}
}

func walkBucketFiles(client *s3.Client, params *s3.ListObjectsV2Input) {
	truncatedListing := true
	for truncatedListing {
		resp, err := client.ListObjectsV2(context.TODO(), params)
		if err != nil {
			exitErrorf("Unable to list items in bucket %q, %v", aws.ToString(params.Bucket), err)
		}
		for _, listObject := range resp.Contents {
			key := aws.ToString(listObject.Key)
			size := aws.ToInt64(listObject.Size)

			lastModified := aws.ToTime(listObject.LastModified)
			const yearsExpiry = 1
			expiry := time.Now().AddDate(-yearsExpiry, 0, 0)
			if lastModified.Before(expiry) {
				// Skip file, it's too old to bother syncing.
				continue
			} else {
				handleFile(client, aws.ToString(params.Bucket), key, size)
			}
		}
		params.ContinuationToken = resp.NextContinuationToken
		truncatedListing = *resp.IsTruncated
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
