package main_test

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
)

// Integration testing for RabbitMQ as it is only used service by producer-service

type IntegrationTestSuite struct {
	suite.Suite
	rabbitMQ *amqp.Connection
	channel  *amqp.Channel
}

func (s *IntegrationTestSuite) SetupSuite() {
	var err error

	rabbitMQConnURL := "amqp://guest:guest@localhost:5672/"

	s.rabbitMQ, err = amqp.Dial(rabbitMQConnURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	s.channel, err = s.rabbitMQ.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}

	// Declare test queue
	_, err = s.channel.QueueDeclare(
		"test_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}
}
