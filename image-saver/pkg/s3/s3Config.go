package s3Adapter

import (
	"os"
)

type S3Config struct {
	Bucket          string
	AccessKeyId     string
	SecretAccessKey string
	Endpoint        string
}

func GetS3Config() S3Config {
	return S3Config{
		os.Getenv("Bucket"),
		os.Getenv("AccessKeyId"),
		os.Getenv("SecretAccessKey"),
		os.Getenv("Endpoint"),
	}
}
