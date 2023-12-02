package services

import (
	"encoding/json"
	"retreival/models"
	"retreival/utils"

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

func (rmq *RabbitMQService) PublishFileData(fileData *models.FileData, queueName string) error {
	fileDataJSON, err := json.Marshal(fileData)
	if err != nil {
		rmq.log.Error("Failed to marshal file data to JSON", zap.Error(err))
		return err
	}

	err = rmq.publishToQueue(fileDataJSON, queueName)
	if err != nil {
		rmq.log.Error("Failed to publish file data", zap.Error(err))
		return err
	}
	rmq.log.Info("File data published successfully", zap.String("QueueName", queueName))
	return nil
}

func (rmq *RabbitMQService) publishToQueue(message []byte, queueName string) error {
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

func (rmq *RabbitMQService) ConsumeQueue(queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := rmq.ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		rmq.log.Error("Failed to register a consumer", zap.Error(err))
		return nil, err
	}
	return msgs, nil
}
