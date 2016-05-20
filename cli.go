package main

import (
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/mitchellh/goamz/aws"
)

const (
	NAME    = "PipelineDeals S3 asset uploader"
	LICENSE = "Licensed under the MIT license"
	VERSION = "0.5.0"
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

	q := make(chan *Upload, *maxQueueSize)
	var wg sync.WaitGroup

	for i := 0; i < *maxWorkers; i++ {
		worker := NewWorker(i, q)
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
			q <- upload
		}()
	}

	wg.Wait()

	log.Println("All done!")
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
	log.Println("Copyright © 2016 PipelineDeals")
	log.Println(LICENSE)
	log.Println("")
}
