package database

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/vatsal3003/viswals/internal/consts"
	"go.uber.org/zap"
)

// Database struct contains postgresql and redis connection
type Database struct {
	PgDB    *sql.DB
	RedisDB *redis.Client
}

// New will initialize postgres and redis database connection and put them into struct and return it
func New(logger *zap.Logger) (*Database, error) {
	// Initialize postgresql database connection
	pgDB, err := sql.Open("postgres", os.Getenv(consts.PostgresConnURL))
	if err != nil {
		logger.Error("failed to connect postgresql database:" + err.Error())
		return nil, err
	}

	// Configure postgres connection
	pgDB.SetMaxOpenConns(20)
	pgDB.SetMaxIdleConns(2)

	// Ping postgresql database to test the connection
	err = pgDB.Ping()
	if err != nil {
		logger.Error("failed to ping postgres database connection:" + err.Error())
		return nil, err
	}

	// Fetch the connection options by parsing the redis connection url
	redisConnOptions, err := redis.ParseURL(os.Getenv(consts.RedisConnURL))
	if err != nil {
		logger.Error("failed to parse redis connection url:" + err.Error())
		return nil, err
	}

	// Initialize the redis database connection
	redisDB := redis.NewClient(redisConnOptions)

	// Ping redis database to test the connection
	status := redisDB.Ping(context.Background())
	if status.Err() != nil {
		logger.Error("failed to ping redis database connection:" + status.Err().Error())
		return nil, err
	}

	return &Database{
		PgDB:    pgDB,
		RedisDB: redisDB,
	}, nil
}

// Migrate will run migration scripts
func (db *Database) Migrate(logger *zap.Logger) error {
	logger.Info("database migration initialized")

	dbDriver, err := postgres.WithInstance(db.PgDB, &postgres.Config{})
	if err != nil {
		logger.Error("failed to get database driver for migration:" + err.Error())
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", dbDriver)
	if err != nil {
		logger.Error("failed to create new migrate instance:" + err.Error())
		return err
	}

	err = m.Down()
	if err != nil && err != migrate.ErrNoChange {
		logger.Error("failed to apply down migrations:" + err.Error())
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Error("failed to apply up migrations:" + err.Error())
		return err
	}

	logger.Info("database migration completed successfully")

	return nil
}

func (db *Database) Close() {
	if db.PgDB != nil {
		err := db.PgDB.Close()
		if err != nil {
			log.Println("ERROR failed to close postgres database connection")
		}
	}

	if db.RedisDB != nil {
		err := db.RedisDB.Close()
		if err != nil {
			log.Println("ERROR failed to close redis database connection:" + err.Error())
			return
		}
	}
}
