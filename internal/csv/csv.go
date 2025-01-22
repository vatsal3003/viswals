package csv

import (
	"encoding/csv"
	"errors"
	"io"
	"os"

	"github.com/vatsal3003/viswals/internal/database"
	"github.com/vatsal3003/viswals/internal/rabbitmq"
	"go.uber.org/zap"
)

// IngestCSV will read from csv and publish message to rabbitmq
func IngestCSV(logger *zap.Logger, rmq *rabbitmq.RabbitMQ) error {
	csvFile, err := os.Open("/users.csv")
	// csvFile, err := os.Open("../../csvs/demo.csv")
	if err != nil {
		logger.Error("failed to open csv file to ingest data:" + err.Error())
		return err
	}
	defer csvFile.Close()

	// Initialize and configure csv reader
	csvReader := csv.NewReader(csvFile)

	csvReader.Comment = '#'
	csvReader.ReuseRecord = true // Keeping it true to reuse the previous slice for storing the new record for better performance

	// Read headers
	row, err := csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("csv file is empty")
		}
		return err
	}

	csvReader.FieldsPerRecord = len(row) // expected fields per row

	// Publish message to rabbitmq
	rmq.Publish(logger, csvReader)

	return nil
}

// DigestCSV will consume messages from rabbitmq
func DigestCSV(logger *zap.Logger, rmq *rabbitmq.RabbitMQ, db *database.Database) {
	usersChan := make(chan []byte, 50)

	// Start consuming messages from rabbitmq
	rmq.Consume(logger, db, usersChan)
}
