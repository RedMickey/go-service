package main

import (
	"database/sql"
	"fmt"
	"image-service/api/routers"
	"image-service/pkg/core"
	"image-service/pkg/dbAdapter"
	"image-service/pkg/rmqAdapter"
	"image-service/pkg/s3Adapter"
	"image-service/pkg/utils"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"

	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		panic(err)
	}

	db, err := databaseConnection()

	if err != nil {
		panic(err)
	}

	imageRepository := dbAdapter.NewImageRepository(db)

	s3Client, uploader, bucketName, err := s3Connection()

	if err != nil {
		panic(err)
	}

	amqpConnection, rmqChannel, queue, err := rmqConnection()
	if err != nil {
		panic(err)
	}
	defer amqpConnection.Close()
	defer rmqChannel.Close()

	rmqAdapter := rmqAdapter.NewRmqAdapter(rmqChannel, queue)
	s3Adapter := s3Adapter.NewS3Adapter(s3Client, uploader, bucketName, rmqAdapter)
	imageService := core.NewImageService(imageRepository, s3Adapter, "http://localhost:3000")

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	api := app.Group("/api")
	routers.ImageRouter(api, imageService)

	app.Listen(":3000")
}

func databaseConnection() (*sql.DB, error) {
	dbConfig := dbAdapter.GetPgDbConfig()
	psqlconn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Dbname,
	)

	fmt.Println(psqlconn)

	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}

func s3Connection() (*s3.S3, *s3manager.Uploader, string, error) {
	s3Config := s3Adapter.GetS3Config()

	awsConfig := aws.NewConfig()
	awsConfig.Endpoint = &s3Config.Endpoint
	awsConfig.Region = aws.String("us-east-2")
	awsConfig.S3ForcePathStyle = utils.BoolPointer(true)
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

func rmqConnection() (*amqp.Connection, *amqp.Channel, *amqp.Queue, error) {
	rmqConfig := rmqAdapter.GetRmqConfig()

	connectionRabbitMQ, err := amqp.Dial(rmqConfig.RMQUrl)
	if err != nil {
		return nil, nil, nil, err
	}

	rmqChannel, err := connectionRabbitMQ.Channel()
	if err != nil {
		return nil, nil, nil, err
	}

	queue, err := rmqChannel.QueueDeclare(
		rmqConfig.RMQQueueName, // name
		false,                  // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)

	return connectionRabbitMQ, rmqChannel, &queue, nil
}
