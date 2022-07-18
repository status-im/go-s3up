package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/apsdehal/go-logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	target   string
	acl      string
	bucket   string
	region   string
	endpoint string
	keyid    string
	secret   string
	threads  int
	debug    bool
)

const helpMessage string = `
This is a simple S3-compatible upload CLI tool.

`

func envVar(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func flagsInit() {
	defaultUsage := flag.Usage
	flag.Usage = func() {
		fmt.Printf(strings.Trim(helpMessage, "\t "))
		defaultUsage()
	}

	flag.StringVar(&target, "target", "", "Path of file to upload.")
	flag.StringVar(&acl, "acl", "private", "Type of permission for file.")
	flag.StringVar(&bucket, "bucket", "", "Name of bucket to upload to.")
	flag.StringVar(&region, "region", envVar("AWS_DEFAULT_REGION", "ams3"), "Name of region to upload to.")
	flag.StringVar(&endpoint, "endpoint", envVar("AWS_DEFAULT_ENDPOINT", "ams3.digitaloceanspaces.com"), "S3 API endpoint.")
	flag.StringVar(&keyid, "keyid", envVar("AWS_ACCESS_KEY_ID", ""), "API key ID.")
	flag.StringVar(&secret, "secret", envVar("AWS_SECRET_ACCESS_KEY", ""), "API secret key.")
	flag.IntVar(&threads, "threads", 20, "Number of concurrent threads used for upload.")
	flag.BoolVar(&debug, "debug", false, "Show debug log messages.")
	flag.Parse()
}

func logInit() *logger.Logger {
	log, err := logger.New("go-s3up", 1, os.Stderr)
	if err != nil {
		panic(err) // Check for error
	}
	if debug {
		log.SetLogLevel(logger.DebugLevel)
	}
	log.SetFormat("[%{module}] %{level}: %{message}")
	return log
}

func main() {
	flagsInit()
	log := logInit()

	if len(keyid) == 0 {
		log.Warning("Provide -keyid flag or AWS_ACCESS_KEY_ID env var")
		os.Exit(1)
	}
	if len(secret) == 0 {
		log.Warning("Provide -secret flag or AWS_SECRET_ACCESS_KEY env var")
		os.Exit(1)
	}

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(keyid, secret, ""),
		Endpoint:    aws.String(endpoint),
		Region:      aws.String(region),
	}

	log.DebugF("Connecting to: %s (region: %s)", endpoint, region)

	newSession := session.New(s3Config)
	uploader := s3manager.NewUploader(newSession)

	file, err := os.Open(target)
	if err != nil {
		log.ErrorF("Failed to open file: %v", err)
		os.Exit(1)
	}

	result, err := uploader.Upload(&s3manager.UploadInput{
		ACL:    aws.String(acl),
		Bucket: aws.String(bucket),
		Key:    aws.String(target),
		Body:   file,
	}, func(u *s3manager.Uploader) {
		u.Concurrency = threads
	})
	if err != nil {
		log.ErrorF("Failed to upload file: %v", err)
		os.Exit(1)
	}

	log.DebugF("ETag: %s", aws.StringValue(result.ETag))
	log.InfoF("Uploaded: %s", aws.StringValue(&result.Location))
}
