package main

import (
	"os"
	"os/signal"

	_ "github.com/lib/pq"
	"github.com/vatsal3003/viswals/internal/csv"
	"github.com/vatsal3003/viswals/internal/logger"
	"github.com/vatsal3003/viswals/internal/rabbitmq"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger := logger.New()
	defer logger.Sync()

	logger.Info("producer service")

	// Initialize rabbitmq
	rmq, err := rabbitmq.New(logger, &rabbitmq.Options{
		Arguments:  nil,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	})
	if err != nil {
		return
	}

	// Gracefully shutdown application
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	go func(rmq *rabbitmq.RabbitMQ, logger *zap.Logger) {
		<-interruptChan
		// rmq.CloseResources()
		logger.Info("resources cleaned")
	}(rmq, logger)

	// Start reading csv and send message to rabbitmq
	csv.IngestCSV(logger, rmq)

	logger.Info("csv ingested successfully")
}
