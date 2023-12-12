package rmqAdapter

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RmqAdapter struct {
	rmqChannel *amqp.Channel
	queue      *amqp.Queue
}

func NewRmqAdapter(rmqChannel *amqp.Channel, queue *amqp.Queue) *RmqAdapter {
	return &RmqAdapter{
		rmqChannel,
		queue,
	}
}

func (rmq *RmqAdapter) PublishToQueue(v interface{}) error {
	bytes, err := json.Marshal(v)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = rmq.rmqChannel.PublishWithContext(
		ctx,
		"",             // exchange
		rmq.queue.Name, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bytes,
		})

	if err != nil {
		return err
	}

	return nil
}
