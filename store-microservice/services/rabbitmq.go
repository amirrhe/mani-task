package services

import (
	"store/utils"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQService struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	log  *zap.Logger
}

func NewRabbitMQService(conn *amqp.Connection, ch *amqp.Channel) *RabbitMQService {
	log := utils.GetLogger()
	return &RabbitMQService{conn, ch, log}
}

func (rmq *RabbitMQService) ConsumeQueue(queueName string) (<-chan amqp.Delivery, error) {
	chanelRabbit, _ := rmq.conn.Channel()
	msgs, err := chanelRabbit.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		rmq.log.Warn("Failed to register a consumer", zap.Error(err))
		_, err := chanelRabbit.QueueDeclare(
			queueName, // Queue name
			true,      // Durable
			false,     // Delete when unused
			false,     // Exclusive
			false,     // No-wait
			nil,       // Arguments
		)
		if err != nil {
			return nil, err
		}
		msgs, _ := chanelRabbit.Consume(
			queueName, // queue
			"",        // consumer
			true,      // auto-ack
			false,     // exclusive
			false,     // no-local
			false,     // no-wait
			nil,       // args
		)
		return msgs, nil

	}
	return msgs, nil
}

func (rmq *RabbitMQService) PublishToQueue(message []byte, queueName string) error {
	// Declare the queue
	q, err := rmq.ch.QueueDeclare(
		queueName, // Name of the queue
		true,      // Durable
		false,     // Delete when unused
		false,     // Exclusive
		false,     // No-wait
		nil,       // Arguments
	)
	if err != nil {
		return err
	}

	err = rmq.ch.Publish(
		"",     // Exchange
		q.Name, // Routing key (queue name)
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
	if err != nil {
		rmq.log.Error("Failed to publish message to queue", zap.Error(err), zap.String("QueueName", queueName))
		return err
	}
	rmq.log.Info("Message published to queue successfully", zap.String("QueueName", queueName))
	return nil
}
