package services_test

import (
	"fmt"
	"os"
	"retreival/models"
	"retreival/services"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

const (
	testQueueName = "testQueue"
)

func LoadEnv() {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
}

func TestRabbitMQService_PublishFileData(t *testing.T) {
	LoadEnv()
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		t.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	rabbitMQService := services.NewRabbitMQService(conn, ch)

	fileData := &models.FileData{
		FileName: "test.txt",
	}

	err = rabbitMQService.PublishFileData(fileData, testQueueName)
	assert.NoError(t, err)
	defer ch.QueueDelete(testQueueName, false, false, false)
}

func TestRabbitMQService_ConsumeQueue(t *testing.T) {
	LoadEnv()
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		t.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	rabbitMQService := services.NewRabbitMQService(conn, ch)

	_, err = ch.QueueDeclare(
		testQueueName, // Name of the test queue
		true,          // Durable
		false,         // Delete when unused
		false,         // Exclusive
		false,         // No-wait
		nil,           // Arguments
	)
	if err != nil {
		t.Fatal("Failed to declare test queue:", err)
	}
	defer ch.QueueDelete(testQueueName, false, false, false)

	msgs, err := rabbitMQService.ConsumeQueue(testQueueName)
	assert.NoError(t, err)

	err = ch.Publish(
		"",            // Exchange
		testQueueName, // Routing key (queue name)
		false,         // Mandatory
		false,         // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(`{"ID": 1, "Name": "test.txt"}`),
		},
	)
	assert.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.NotNil(t, msg)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message from test queue")
	}
}
