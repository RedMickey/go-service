package main

import (
	"encoding/json"
	"fmt"
	"log"

	"image-saver/pkg/imageProcessor"
	s3Adapter "image-saver/pkg/s3"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ImageQueueMessageData struct {
	OriginalImageName string   `json:"originalImageName"`
	SaveName          string   `json:"saveName"`
	SaveFormats       []string `json:"saveFormats"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	err := godotenv.Load()

	failOnError(err, "")

	amqpConnection, err := rmqConnection()

	failOnError(err, "")

	defer amqpConnection.Close()

	rmqChannel, err := amqpConnection.Channel()

	failOnError(err, "")

	defer rmqChannel.Close()

	queue, err := rmqChannel.QueueDeclare(
		os.Getenv("RMQQueueName"), // name
		false,                     // durable
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		nil,                       // arguments
	)

	failOnError(err, "")

	messages, err := rmqChannel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)

	failOnError(err, "")

	s3Client, uploader, bucketName, err := s3Connection()

	failOnError(err, "")

	s3Adapter := s3Adapter.NewS3Adapter(s3Client, uploader, bucketName)
	imgProcessor := imageProcessor.NewImageProcessor()

	var forever chan struct{}

	go func() {
		for message := range messages {
			messageStr := string(message.Body)
			fmt.Println("received a message:", messageStr)

			var imageQueueMessageData ImageQueueMessageData

			err := json.Unmarshal(message.Body, &imageQueueMessageData)

			if err == nil {
				var originalImage []byte = []byte{}
				originalImage, err = s3Adapter.GetFile(imageQueueMessageData.OriginalImageName)

				if err == nil {
					var processedImg *imageProcessor.ImageData

					for _, format := range imageQueueMessageData.SaveFormats {
						fmt.Println(format)
						processedImg, err = imgProcessor.ConvertImage(
							originalImage,
							imageQueueMessageData.OriginalImageName,
							imageQueueMessageData.SaveName,
							format,
						)

						if err == nil {
							err = s3Adapter.SaveImageFormat(*processedImg)
						}

						if err != nil {
							break
						}
					}

					if err == nil {
						err = s3Adapter.DeleteFile(imageQueueMessageData.OriginalImageName)

						if err != nil {
							fmt.Println(fmt.Sprintf("%s: %s", "Failed to delete the original image file", err))
						}
					}
				}
			}

			if err != nil {
				fmt.Println(err)
				message.Reject(false)
			} else {
				message.Ack(false)
				fmt.Println("Image has been processed successfully")
			}
		}
	}()

	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func s3Connection() (*s3.S3, *s3manager.Uploader, string, error) {
	s3Config := s3Adapter.GetS3Config()

	awsConfig := aws.NewConfig()
	awsConfig.Endpoint = &s3Config.Endpoint
	awsConfig.Region = aws.String("us-east-2")
	awsConfig.S3ForcePathStyle = aws.Bool(true)
	awsConfig.Credentials = credentials.NewStaticCredentials(
		s3Config.AccessKeyId,
		s3Config.SecretAccessKey,
		"",
	)

	sess := session.Must(session.NewSession(awsConfig))
	uploader := s3manager.NewUploader(sess)
	s3Client := s3.New(sess, awsConfig)

	return s3Client, uploader, s3Config.Bucket, nil
}

func rmqConnection() (*amqp.Connection, error) {
	amqpServerURL := os.Getenv("RMQUrl")

	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	if err != nil {
		return nil, err
	}

	return connectRabbitMQ, nil
}
