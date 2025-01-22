package consts

var (
	// ContentType constants
	ContentTypeGob = "application/x-gob"

	// LogLevel constants
	LogLevelDebug = "DEBUG"

	// Response constant
	StatusSuccess = "success"
	StatusError   = "error"

	// env constants
	LogLevel          = "LOG_LEVEL"
	MigrateDatabase   = "MIGRATE_DB"
	ConsumerPort      = "CONSUMER_PORT"
	PostgresConnURL   = "POSTGRES_CONN_URL"
	RedisConnURL      = "REDIS_CONN_URL"
	RabbitMQConnURL   = "RABBITMQ_CONN_URL"
	RabbitMQQueueName = "RABBITMQ_QUEUE_NAME"
)
