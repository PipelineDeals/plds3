package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

const (
	NAME    = "PipelineDeals S3 asset uploader"
	LICENSE = "Licensed under the MIT license"
	VERSION = "0.3.0"
)

var (
	bucket       = flag.String("b", "assets.pipelinedeals.com", "AWS S3 bucket to upload assets to")
	directory    = flag.String("d", "", "Directory of assets to upload")
	key          = flag.String("k", "", "Key to preface to asset")
	maxWorkers   = flag.Int("w", 50, "Max number of workers to start")
	maxQueueSize = flag.Int("q", 1000, "Max size of upload queue")
	help         = flag.Bool("help", false, "You're looking at it")
	h            = flag.Bool("h", false, "You're looking at it")
	version      = flag.Bool("v", false, "Show version")
)

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

	err = b.PutReaderHeader(relativePath, file, stat.Size(), headers, s3.ACL("public-read"))

	fmt.Printf("Path: %s\n", relativePath)

	if err != nil {
		log.Fatal(err)
	}

}

type Worker struct {
	id          int
	uploadQueue chan *Upload
}

func NewWorker(id int, uploadQueue chan *Upload) Worker {
	return Worker{
		id:          id,
		uploadQueue: uploadQueue,
	}
}

func (w Worker) start() {
	go func() {
		for {
			select {
			case upload := <-w.uploadQueue:
				upload.Put()
				upload.WaitGroup.Done()
			}
		}
	}()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

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

	_, err := aws.EnvAuth()

	if err != nil {
		log.Fatal(err)
	}

	uploadQueue := make(chan *Upload, *maxQueueSize)
	var wg sync.WaitGroup

	for i := 0; i < *maxWorkers; i++ {
		worker := NewWorker(i, uploadQueue)
		worker.start()
	}

	dir := path.Clean(*directory)
	publicFileList, err := GetFileList(dir)

	if err != nil {
		log.Fatal(err)
	}

	for _, path := range publicFileList {
		upload := &Upload{
			Path:      path,
			Dir:       dir,
			Bucket:    *bucket,
			PrefixKey: *key,
			WaitGroup: &wg,
		}

		wg.Add(1)

		go func() {
			uploadQueue <- upload
		}()
	}

	wg.Wait()
}

func GetFileList(dir string) (fileList []string, err error) {
	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Looks like %s isn't a valid app directory\n", path)
		}

		if fi.Mode().IsRegular() {
			fileList = append(fileList, path)
		}
		return nil
	})

	return fileList, err
}

func printHelp() {
	log.Println("-b\t\tAWS S3 target bucket, default: assets.pipelinedeals.com")
	log.Println("-d\t\tDirectory of assets to upload, default: ''")
	log.Println("-k\t\tKey to preface the asset with, default: ''")
	log.Println("-w\t\tMax number of workers to start, default: 50")
	log.Println("-q\t\tKey Max size of upload queue, default: 1000")
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
