package rabbitmq

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	channel    *amqp.Channel
	exchange   string
	routingKey string
}

func NewRabbitPublisher(conn *amqp.Connection, exchange, routingKey string) (*RabbitPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		exchange,
		"topic", // или "direct"
		true,    // durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &RabbitPublisher{
		channel:    ch,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (p *RabbitPublisher) Publish(ctx context.Context, body json.RawMessage) error {
	return p.channel.PublishWithContext(ctx,
		p.exchange,
		p.routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
