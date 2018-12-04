package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type downloader struct {
	*s3manager.Downloader
	bucket, dir string
}

func (d *downloader) eachPage(page *s3.ListObjectsOutput, more bool) bool {
	for _, obj := range page.Contents {
		d.downloadToFile(*obj.Key)
	}

	return true
}

func (d *downloader) downloadToFile(key string) {
	// Create the directories in the path
	file := filepath.Join(d.dir, key)
	if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
		log.Panic(err)
	}

	// Set up the local file
	fd, err := os.Create(file)
	if err != nil {
		log.Panic(err)
	}
	defer fd.Close()

	// Download the file using the AWS SDK for Go
	log.Printf("Downloading s3://%s/%s to %s...\n", d.bucket, key, file)
	params := &s3.GetObjectInput{Bucket: &d.bucket, Key: &key}
	d.Download(fd, params)
}

func updateServerS3() {
	manager := s3manager.NewDownloader(awsSession)
	d := downloader{bucket: Config.S3Bucket, dir: Config.S3Folder, Downloader: manager}

	params := &s3.ListObjectsInput{
		Bucket: aws.String(Config.S3Bucket),
		Prefix: aws.String(Config.S3BucketPrefix),
	}

	s3Service.ListObjectsPages(params, d.eachPage)
}
