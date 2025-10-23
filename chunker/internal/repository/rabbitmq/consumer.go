package rabbitmq

import (
	"chunker/internal/domain/entity"
	"chunker/internal/domain/usecase"
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ChunkerConsumer struct {
	channel     *amqp.Channel
	exchange    string
	routingKey  string
	queue       string
	UseCase     *usecase.ChunkerUseCase
	prefetchCnt int
}

func NewChunkerConsumer(conn *amqp.Connection, exchange, routingKey, queue string, uc *usecase.ChunkerUseCase) (*ChunkerConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	consumer := &ChunkerConsumer{
		channel:     ch,
		exchange:    exchange,
		routingKey:  routingKey,
		queue:       queue,
		UseCase:     uc,
		prefetchCnt: 1,
	}

	_, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if err := ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	if err := ch.Qos(consumer.prefetchCnt, 0, false); err != nil {
		return nil, err
	}

	return consumer, nil
}

func (c *ChunkerConsumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("ChunkerConsumer shutting down")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("RabbitMQ channel closed")
				return nil
			}

			var job entity.Job
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				log.Println("failed to unmarshal job:", err)
				msg.Nack(false, false)
				continue
			}

			log.Println(job)

			go func(job entity.Job, msg amqp.Delivery) {
				if err := c.UseCase.ProcessJob(ctx, &job); err != nil {
					log.Printf("failed to process job %s: %v\n", job.JobID, err)
					msg.Nack(false, true)
					return
				}
				msg.Ack(false)
			}(job, msg)
		}
	}
}
