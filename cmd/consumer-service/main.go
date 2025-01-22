package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/vatsal3003/viswals/internal/consts"
	"github.com/vatsal3003/viswals/internal/csv"
	"github.com/vatsal3003/viswals/internal/database"
	"github.com/vatsal3003/viswals/internal/logger"
	"github.com/vatsal3003/viswals/internal/rabbitmq"
	"github.com/vatsal3003/viswals/internal/usersapi"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger := logger.New()
	defer logger.Sync()

	// Initialize database
	db, err := database.New(logger)
	if err != nil {
		return
	}

	// Migrate database
	if os.Getenv(consts.MigrateDatabase) == "true" {
		err := db.Migrate(logger)
		if err != nil {
			return
		}
	}

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

	// Start consuming messages from rabbitmq
	go csv.DigestCSV(logger, rmq, db)

	// Initialize users api
	api := usersapi.New(db, logger)

	// Define routes
	api.InitRoutes()

	// Define server
	server := &http.Server{
		Addr:    os.Getenv(consts.ConsumerPort),
		Handler: http.DefaultServeMux,
	}

	// Gracefully shutdown application
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	go func(rmq *rabbitmq.RabbitMQ, logger *zap.Logger) {
		<-interruptChan
		rmq.CloseResources()
		db.Close()
		err = server.Shutdown(context.Background())
		if err != nil {
			logger.Error("failed to shutdown http server:" + err.Error())
		}
		logger.Info("resources cleaned")
	}(rmq, logger)

	// Start http server
	err = server.ListenAndServe()
	if err != nil {
		logger.Error("failed to start server:" + err.Error())
	}
}
