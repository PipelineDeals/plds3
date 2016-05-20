package main

import (
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

var MaxAttempts = 5

type Upload struct {
	Path      string
	Dir       string
	Bucket    string
	PrefixKey string
	WaitGroup *sync.WaitGroup
}

func (u *Upload) RelativePath() string {
	relativePath := strings.TrimPrefix(u.Path, u.Dir+"/")

	if u.PrefixKey != "" {
		relativePath = u.PrefixKey + "/" + relativePath
	}
	return relativePath
}

func (u *Upload) FileType() string {
	ext := filepath.Ext(u.Path)
	fileType := mime.TypeByExtension(ext)
	return fileType
}

func (u *Upload) Put() {
	auth, _ := aws.EnvAuth()
	client := s3.New(auth, aws.USEast)
	b := client.Bucket(u.Bucket)

	file, err := os.Open(u.Path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	stat, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	headers := map[string][]string{
		"Content-Length": {strconv.FormatInt(stat.Size(), 10)},
		"Content-Type":   {u.FileType()},
		"Cache-Control":  {"max-age=31104000"},
	}

	relativePath := u.RelativePath()

	attempt := 1
	for {
		fmt.Printf("[%d] Path: %s\n", attempt, relativePath)

		err = b.PutReaderHeader(relativePath, file, stat.Size(), headers, s3.ACL("public-read"))
		if err == nil || attempt >= MaxAttempts {
			break
		}

		time.Sleep(time.Duration(attempt) * time.Second)
		attempt += 1
		file.Seek(0 ,0)
	}

	if err != nil {
		log.Fatal(err)
	}

}
