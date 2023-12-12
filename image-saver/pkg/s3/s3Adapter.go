package s3Adapter

import (
	"bytes"
	"image-saver/pkg/imageProcessor"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Adapter struct {
	s3Client   *s3.S3
	s3Uploader *s3manager.Uploader
	bucketName string
}

func NewS3Adapter(s3Client *s3.S3, s3Uploader *s3manager.Uploader, bucketName string) *S3Adapter {
	return &S3Adapter{
		s3Client,
		s3Uploader,
		bucketName,
	}
}

func (s *S3Adapter) GetFile(name string) ([]byte, error) {
	getObjectOutput, err := s.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(name),
	})

	if err != nil {
		return []byte{}, err
	}

	defer getObjectOutput.Body.Close()

	file, err := io.ReadAll(getObjectOutput.Body)

	if err != nil {
		return []byte{}, err
	}

	return file, nil
}

func (s *S3Adapter) SaveImageFormat(imageData imageProcessor.ImageData) error {
	imgBuf := bytes.NewBuffer(imageData.File)

	_, err := s.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(imageData.Name),
		Body:   imgBuf,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *S3Adapter) DeleteFile(name string) error {
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(name),
	})

	return err
}
