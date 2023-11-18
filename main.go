package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"os"
	"path/filepath"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg)
	const maxKeys = 1000

	totalQty := 0
	bucket := os.Getenv("BUCKET")
	listInput := s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(maxKeys),
		Prefix:  aws.String(os.Getenv("MAIN_FOLDER")),
	}

	listFiles := getListFiles(client, &listInput)
	for _, key := range listFiles {
		err := handleFile(key)

		if err != nil {
			log.Fatal(err)
		}
		totalQty += 1
	}

	log.Printf("totalQty=%d", totalQty)
}

func handleFile(key string) error {
	savePath := os.Getenv("SAVE_PATH")
	file := fmt.Sprintf("%s/%s", savePath, key)
	dir := filepath.Dir(file)
	parent := filepath.Base(dir)
	log.Printf("file=%s", file)
	log.Printf("dir=%s", dir)
	log.Printf("parent=%s", parent)
	// if file doesn't exist
	// download to location

	return nil
}

func getListFiles(client *s3.Client, params *s3.ListObjectsV2Input) []string {
	var listUrl []string
	truncatedListing := true
	for truncatedListing {
		resp, err := client.ListObjectsV2(context.TODO(), params)
		if err != nil {
			exitErrorf("Unable to list items in bucket %q, %v", aws.ToString(params.Bucket), err)
		}
		for _, object := range resp.Contents {
			key := aws.ToString(object.Key)
			listUrl = append(listUrl, key)
		}
		params.ContinuationToken = resp.NextContinuationToken
		truncatedListing = *resp.IsTruncated

	}

	return listUrl
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
