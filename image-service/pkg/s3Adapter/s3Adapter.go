package s3Adapter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image-service/pkg/core"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var supportedFileTypes map[string]string = map[string]string{
	"jpeg": "jpeg",
	"jpg":  "jpg",
	"png":  "png",
	"webp": "webp",
	"avif": "avif",
}

type QueuePublisher interface {
	PublishToQueue(v interface{}) error
}

type ImageQueueMessageData struct {
	OriginalImageName string   `json:"originalImageName"`
	SaveName          string   `json:"saveName"`
	SaveFormats       []string `json:"saveFormats"`
}

type s3Adapter struct {
	s3Client       *s3.S3
	s3Uploader     *s3manager.Uploader
	bucketName     string
	queuePublisher QueuePublisher
}

func NewS3Adapter(s3Client *s3.S3, s3Uploader *s3manager.Uploader, bucketName string, queuePublisher QueuePublisher) core.DataStorage {
	return &s3Adapter{
		s3Client,
		s3Uploader,
		bucketName,
		queuePublisher,
	}
}

func (s *s3Adapter) GetFile(name string) ([]byte, error) {
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

func (s *s3Adapter) SaveImageAsync(file []byte, originalImageName string, saveName string, formats []string) error {
	extName := strings.ToLower(strings.Replace(filepath.Ext(originalImageName), ".", "", -1))
	imgBuf := bytes.NewBuffer(file)

	originalImageSaveName := saveName + "-original" + "." + extName

	_, err := s.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(originalImageSaveName),
		Body:   imgBuf,
	})

	if err != nil {
		return err
	}

	var supportedFormatsToSave []string = make([]string, 0)

	for _, format := range formats {
		if _, prs := supportedFileTypes[format]; prs {
			supportedFormatsToSave = append(supportedFormatsToSave, format)
		} else {
			fmt.Println(fmt.Sprintf("Формат %s не поддерживается", format))
		}
	}

	s.saveImageFormatAsync(originalImageSaveName, saveName, supportedFormatsToSave)

	return nil
}

func (s *s3Adapter) SaveImage(file []byte, name string, formats []string) error {
	for _, format := range formats {
		if _, prs := supportedFileTypes[format]; prs {
			s.saveImageFormat(file, name, format)
		} else {
			fmt.Println(fmt.Sprintf("Формат %s не поддерживается", format))
		}
	}

	return nil
}

func (s *s3Adapter) DeleteImage(name string, formats []string) error {
	for _, format := range formats {
		err := s.DeleteFile(name + "." + format)

		if err != nil {
			fmt.Println(fmt.Sprintf("%s: %s", "Failed to delete the original image file", err))
		}
	}

	return nil
}

func (s *s3Adapter) DeleteFile(name string) error {
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(name),
	})

	return err
}

func (s *s3Adapter) saveImageFormat(file []byte, name string, format string) error {
	imgBuf := bytes.NewBuffer(file)
	imgDecoded, _, err := image.Decode(imgBuf)

	if err != nil {
		return err
	}

	var encodedBuf bytes.Buffer
	w := bufio.NewWriter(&encodedBuf)

	switch format {
	case "jpg", "jpeg":
		err = jpeg.Encode(w, imgDecoded, nil)
	case "png":
		png.Encode(w, imgDecoded)
	default:
		err = errors.New(fmt.Sprintf("Формат %s не поддерживается", format))
	}

	if err != nil {
		return err
	}

	_, err = s.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(name + "." + format),
		Body:   &encodedBuf,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *s3Adapter) saveImageFormatAsync(originalImageName string, saveName string, saveFormats []string) error {
	imageQueueMessageData := ImageQueueMessageData{
		OriginalImageName: originalImageName,
		SaveName:          saveName,
		SaveFormats:       saveFormats,
	}

	s.queuePublisher.PublishToQueue(imageQueueMessageData)

	return nil
}
