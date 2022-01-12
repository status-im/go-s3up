package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var l *log.Logger

var (
	target   string
	acl      string
	bucket   string
	region   string
	endpoint string
	keyid    string
	secret   string
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
	flag.Parse()
}

func main() {
	l = log.New(os.Stderr, "", log.Lshortfile)

	flagsInit()

	if len(keyid) == 0 {
		l.Println("Provide -keyid flag or AWS_ACCESS_KEY_ID env var")
		os.Exit(1)
	}
	if len(secret) == 0 {
		l.Println("Provide -secret flag or AWS_SECRET_ACCESS_KEY env var")
		os.Exit(1)
	}

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(keyid, secret, ""),
		Endpoint:    aws.String(endpoint),
		Region:      aws.String(region),
	}

	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

	file, err := os.Open(target)
	if err != nil {
		l.Println("Failed to open file: ", err)
		os.Exit(1)
	}

	object := s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(target),
		ACL:    aws.String(acl),
		Body:   file,
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		fmt.Println(err.Error())
	}
}
