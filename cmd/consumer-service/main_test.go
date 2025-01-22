package main_test

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Integration testing for PostgreSQL, Redis and RabbitMQ as it is services used by consumer-service

type IntegrationTestSuite struct {
	suite.Suite
	pgDB     *sql.DB
	redisDB  *redis.Client
	rabbitMQ *amqp.Connection
	channel  *amqp.Channel
}

func (s *IntegrationTestSuite) SetupSuite() {
	var err error

	pgConnStr := "host=localhost port=5432 user=postgres password=admin dbname=postgres sslmode=disable"

	s.pgDB, err = sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	_, err = s.pgDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(100)
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create test table: %v", err)
	}

	redisAddr := "localhost:6379"

	s.redisDB = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	// Test Redis connection
	ctx := context.Background()
	_, err = s.redisDB.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// RabbitMQ Setup
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

func (s *IntegrationTestSuite) TearDownSuite() {
	// Clean up PostgreSQL
	_, err := s.pgDB.Exec("DROP TABLE IF EXISTS test_users")
	if err != nil {
		log.Printf("Failed to drop test table: %v", err)
	}
	s.pgDB.Close()

	// Clean up Redis
	s.redisDB.Close()

	// Clean up RabbitMQ
	s.channel.QueueDelete("test_queue", false, false, false)
	s.channel.Close()
	s.rabbitMQ.Close()
}

func (s *IntegrationTestSuite) TestPostgreSQL() {
	// Test PostgreSQL operations
	s.Run("PostgreSQL CRUD Operations", func() {
		result, err := s.pgDB.Exec(
			"INSERT INTO test_users (name, email) VALUES ($1, $2)",
			"Test User",
			"test@example.com",
		)
		assert.NoError(s.T(), err)

		_, err = result.LastInsertId()
		if err != nil {
			// PostgreSQL doesn't support LastInsertId, get ID using RETURNING
			var insertedID int64
			err = s.pgDB.QueryRow(
				"INSERT INTO test_users (name, email) VALUES ($1, $2) RETURNING id",
				"Test User 2",
				"test2@example.com",
			).Scan(&insertedID)
			assert.NoError(s.T(), err)
			assert.Greater(s.T(), insertedID, int64(0))
		}

		// Select
		var name, email string
		err = s.pgDB.QueryRow(
			"SELECT name, email FROM test_users WHERE email = $1",
			"test@example.com",
		).Scan(&name, &email)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), "Test User", name)
	})
}

func (s *IntegrationTestSuite) TestRedis() {
	ctx := context.Background()

	s.Run("Redis Set and Get", func() {
		// Test Set
		err := s.redisDB.Set(ctx, "test_key", "test_value", time.Minute).Err()
		assert.NoError(s.T(), err)

		// Test Get
		val, err := s.redisDB.Get(ctx, "test_key").Result()
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), "test_value", val)
	})

	s.Run("Redis Expiration", func() {
		// Set with expiration
		err := s.redisDB.Set(ctx, "temp_key", "temp_value", time.Second).Err()
		assert.NoError(s.T(), err)

		// Wait for expiration
		time.Sleep(time.Second * 2)

		// Key should be gone
		_, err = s.redisDB.Get(ctx, "temp_key").Result()
		assert.Error(s.T(), err)
		assert.Equal(s.T(), redis.Nil, err)
	})
}

func (s *IntegrationTestSuite) TestRabbitMQ() {
	s.Run("RabbitMQ Publish and Consume", func() {
		// Publish message
		message := "test message"
		err := s.channel.Publish(
			"",           // exchange
			"test_queue", // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(message),
			},
		)
		assert.NoError(s.T(), err)

		// Consume message
		msgs, err := s.channel.Consume(
			"test_queue",
			"",    // consumer
			true,  // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
		assert.NoError(s.T(), err)

		// Wait for message
		select {
		case msg := <-msgs:
			assert.Equal(s.T(), message, string(msg.Body))
		case <-time.After(time.Second * 5):
			s.T().Error("Timeout waiting for message")
		}
	})
}

// Start integration test
func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}
