package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

const (
	NAME    = "PipelineDeals S3 asset uploader"
	LICENSE = "Licensed under the MIT license"
	VERSION = "0.1"
)

var (
	bucket    = flag.String("b", "assets.pipelinedeals.com", "AWS S3 bucket to upload assets to")
	directory = flag.String("d", "./public", "Local assets directory to upload to S3")
	key       = flag.String("k", "", "Key to preface to asset")
	help      = flag.Bool("help", false, "You're looking at it")
	h         = flag.Bool("h", false, "You're looking at it")
	version   = flag.Bool("v", false, "Show version")
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	startupInfo()

	flag.Usage = printHelp

	flag.Parse()

	if *help || *h {
		printHelp()
		os.Exit(0)
	}

	if *version {
		os.Exit(0)
	}

	fileList, err := GetFileList(*directory)

	if err != nil {
		log.Fatal(err)
	}

	for _, path := range fileList {
		err := UploadFile(path, *directory, *bucket, *key)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetFileList(dir string) (fileList []string, err error) {
	err = filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})

	return fileList, err
}

func UploadFile(path string, dir string, bucketName string, prefixKey string) (err error) {
	auth, err := aws.EnvAuth()

	if err != nil {
		return err
	}

	client := s3.New(auth, aws.USEast)
	b := client.Bucket(bucketName)

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	fileType := mime.TypeByExtension(ext)

	headers := map[string][]string{
		"Content-Length": {strconv.FormatInt(stat.Size(), 10)},
		"Content-Type":   {fileType},
	}

	relativePath := strings.TrimPrefix(path, dir+"/")

	if prefixKey != "" {
		relativePath = "/" + prefixKey + "/" + relativePath
	}

	err = b.PutReaderHeader(relativePath, file, stat.Size(), headers, s3.ACL("public-read"))

	if err != nil {
		return err
	}

	fmt.Println(relativePath)

	return nil
}

func printHelp() {
	log.Println("-b\t\tAWS S3 target bucket, default: assets.pipelinedeals.com")
	log.Println("-d\t\tLocal assets directory to upload to S3, default: ./public")
	log.Println("-k\t\tKey to preface the asset with, default: ''")
	log.Println("-v\t\tShow version and license information")
	log.Println("-h\t\tThis help screen")
}

func startupInfo() {
	log.Println(NAME, VERSION)
	log.Println("")
	log.Println("Copyright © 2015 PipelineDeals")
	log.Println(LICENSE)
	log.Println("")
}