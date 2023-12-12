package rmqAdapter

import (
	"os"
)

type RmqConfig struct {
	RMQUrl       string
	RMQQueueName string
}

func GetRmqConfig() RmqConfig {
	return RmqConfig{
		os.Getenv("RMQUrl"),
		os.Getenv("RMQQueueName"),
	}
}
