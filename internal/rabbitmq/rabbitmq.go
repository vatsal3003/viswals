package rabbitmq

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"os"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vatsal3003/viswals/internal/consts"
	"github.com/vatsal3003/viswals/internal/database"
	"github.com/vatsal3003/viswals/internal/encryption"
	"github.com/vatsal3003/viswals/internal/service/userservice"
	"github.com/vatsal3003/viswals/internal/utils"
	"github.com/vatsal3003/viswals/models"
	"go.uber.org/zap"
)

type RabbitMQ struct {
	queue   amqp.Queue
	channel *amqp.Channel
	conn    *amqp.Connection
}

type Options struct {
	Arguments  amqp.Table
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
}

func New(logger *zap.Logger, opts *Options) (*RabbitMQ, error) {
	var rabbitmq = &RabbitMQ{}
	var err error

	rabbitmq.conn, err = amqp.Dial(os.Getenv(consts.RabbitMQConnURL))
	if err != nil {
		logger.Error("failed to connect with rabbitmq:" + err.Error())
		return nil, err
	}

	rabbitmq.channel, err = rabbitmq.conn.Channel()
	if err != nil {
		logger.Error("failed to open rabbitmq connection channel:" + err.Error())
		return nil, err
	}

	rabbitmq.queue, err = rabbitmq.channel.QueueDeclare(
		os.Getenv(consts.RabbitMQQueueName), // name
		opts.Durable,                        // durable
		opts.AutoDelete,                     // delete when unused
		opts.Exclusive,                      // exclusive
		opts.NoWait,                         // no-wait
		opts.Arguments,                      // arguments
	)
	if err != nil {
		logger.Error("failed to declare a queue from connection channel:" + err.Error())
		return nil, err
	}

	return rabbitmq, nil
}

func (rmq *RabbitMQ) Publish(logger *zap.Logger, csvReader *csv.Reader) error {
	// buf is used to hold gob encoded user data
	var buf bytes.Buffer
	user := &models.User{}

	for {
		encoder := gob.NewEncoder(&buf)
		row, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else if errors.Is(err, csv.ErrFieldCount) {
				continue // continue if the row has partial data
			}
		}

		id, _ := strconv.Atoi(row[0])
		puid, _ := strconv.Atoi(row[7])
		if puid != -1 {
			user.ParentUserID = &puid
		} else {
			user.ParentUserID = nil
		}

		user.ID = id
		user.CreatedAt = *utils.MillisToTime(row[4])
		user.DeletedAt = utils.MillisToTime(row[5])
		user.MergedAt = utils.MillisToTime(row[6])
		user.FirstName = row[1]
		user.LastName = row[2]
		user.EmailAddress = row[3]

		err = encoder.Encode(user)
		if err != nil {
			logger.Error("failed to encode user data into gob stream:" + err.Error())
			return err
		}

		err = rmq.channel.PublishWithContext(
			context.Background(), // context
			"",                   // exchange
			rmq.queue.Name,       // key
			false,                // mandatory
			false,                // immediate
			amqp.Publishing{
				ContentType: consts.ContentTypeGob,
				Body:        buf.Bytes(),
			}, // message
		)
		if err != nil {
			logger.Error("failed to publish message:" + err.Error())
			return err
		}

		buf.Reset()
	}

	return nil
}

func (rmq *RabbitMQ) Consume(logger *zap.Logger, db *database.Database, usersChan chan []byte) {
	messages, err := rmq.channel.ConsumeWithContext(
		context.Background(), // context
		rmq.queue.Name,       // queue
		"",                   // consumer
		true,                 // auto acknowledge
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // args
	)
	if err != nil {
		logger.Error("failed to consume gob stream")
	}

	var forever = make(chan struct{})

	var errChan = make(chan error)

	go func(db *database.Database, usersChan chan []byte, errChan chan error) {
		for msg := range usersChan {
			var buf bytes.Buffer
			encoder := gob.NewEncoder(&buf)

			decoder := gob.NewDecoder(bytes.NewReader(msg))

			var user models.User
			err := decoder.Decode(&user)
			if err != nil {
				logger.Error("failed to decode user from gob stream in consumer:" + err.Error())
				errChan <- err
				return
			}

			user.EmailAddress, err = encryption.Encrypt(user.EmailAddress)
			if err != nil {
				logger.Error("failed to encrypt the user email address:" + err.Error())
				errChan <- err
				return
			}

			err = encoder.Encode(user)
			if err != nil {
				errChan <- err
				return
			}

			// Insert into database
			go func(errChan chan error, user models.User) {
				err := userservice.InsertUser(db, &user)
				if err != nil {
					logger.Error("failed to insert user into database:" + err.Error())
					errChan <- err
					return
				}
			}(errChan, user)
			// Insert into cache
			go func(errChan chan error, userID int, buf []byte) {
				err := userservice.InsertUserInKVStore(db, userID, buf)
				if err != nil {
					logger.Error("failed to insert user into cache:" + err.Error())
					errChan <- err
					return
				}
			}(errChan, user.ID, buf.Bytes())

			buf.Reset()
		}
	}(db, usersChan, errChan)

	go func(usersChan chan []byte) {
		for message := range messages {
			if message.ContentType != consts.ContentTypeGob {
				logger.Error("invalid content-type, expected application/x-gob")
				errChan <- err
			}

			usersChan <- message.Body
		}
	}(usersChan)

	select {
	case err := <-errChan:
		logger.Error("failed to consume message:" + err.Error())
		return
	case <-forever: // blocking this function forever
	}
}

func (rmq *RabbitMQ) CloseResources() {
	err := rmq.channel.Close()
	if err != nil {
		log.Println("ERROR failed to close rabbitmq channel:" + err.Error())
		return
	}

	err = rmq.conn.Close()
	if err != nil {
		log.Println("ERROR failed to close rabbitmq connection:" + err.Error())
		return
	}
}
