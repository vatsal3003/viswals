package userservice_test

// import (
// 	"bytes"
// 	"context"
// 	"database/sql"
// 	"encoding/gob"
// 	"testing"
// 	"time"

// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/go-redis/redis"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/vatsal3003/viswals/internal/database"
// 	"github.com/vatsal3003/viswals/internal/encryption"
// 	"github.com/vatsal3003/viswals/internal/service/userservice"
// 	"github.com/vatsal3003/viswals/models"
// )

// type MockRedisClient struct {
// 	mock.Mock
// }

// func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
// 	args := m.Called(ctx, key)
// 	return args.Get(0).(*redis.StringCmd)
// }

// func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
// 	args := m.Called(ctx, key, value, expiration)
// 	return args.Get(0).(*redis.StatusCmd)
// }

// type MockDB struct {
// 	mock.Mock
// }

// func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
// 	mockArgs := m.Called(query, args)
// 	return mockArgs.Get(0).(*sql.Rows), mockArgs.Error(1)
// }

// func TestInsertUser(t *testing.T) {
// 	mockDB := &database.Database{
// 		PgDB: &MockDB{},
// 	}

// 	user := &models.User{
// 		ID:           1,
// 		FirstName:    "John",
// 		LastName:     "Doe",
// 		EmailAddress: "john.doe@example.com",
// 		CreatedAt:    time.Now(),
// 	}

// 	// Simulate successful user insert
// 	mockDB.PgDB.(*MockDB).On("Exec", mock.Anything, mock.Anything).Return(nil)

// 	err := InsertUser(mockDB, user)
// 	assert.NoError(t, err)
// 	mockDB.PgDB.(*MockDB).AssertExpectations(t)
// }

// func TestGetUser_FromCache(t *testing.T) {
// 	mockRedis := &MockRedisClient{}
// 	mockDB := &database.Database{
// 		RedisDB: mockRedis,
// 	}

// 	user := &models.User{
// 		ID:           1,
// 		FirstName:    "John",
// 		LastName:     "Doe",
// 		EmailAddress: "john.doe@example.com",
// 		CreatedAt:    time.Now(),
// 	}

// 	// Encrypt email for comparison
// 	encryptedEmail, _ := encryption.Encrypt(user.EmailAddress)
// 	user.EmailAddress = encryptedEmail

// 	var buf bytes.Buffer
// 	_ = gob.NewEncoder(&buf).Encode(user)

// 	// Mock Redis Get
// 	mockRedis.On("Get", mock.Anything, "users:1").Return(redis.NewStringResult(buf.String(), nil))

// 	retrievedUser, err := GetUser(mockDB, "1")
// 	assert.NoError(t, err)
// 	assert.Equal(t, user.ID, retrievedUser.ID)
// 	assert.Equal(t, "john.doe@example.com", retrievedUser.EmailAddress)

// 	mockRedis.AssertExpectations(t)
// }

// func TestGetUser_FromDB(t *testing.T) {
// 	mockRedis := &MockRedisClient{}
// 	mockDB := &database.Database{
// 		PgDB:    &MockDB{},
// 		RedisDB: mockRedis,
// 	}

// 	user := &models.User{
// 		ID:           1,
// 		FirstName:    "John",
// 		LastName:     "Doe",
// 		EmailAddress: "john.doe@example.com",
// 		CreatedAt:    time.Now(),
// 	}

// 	// Encrypt email for comparison
// 	encryptedEmail, _ := encryption.Encrypt(user.EmailAddress)
// 	user.EmailAddress = encryptedEmail

// 	mockRows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email_address", "created_at", "deleted_at", "merged_at", "parent_user_id"}).
// 		AddRow(user.ID, user.FirstName, user.LastName, user.EmailAddress, user.CreatedAt, nil, nil, nil)

// 	// Mock Redis Get (cache miss)
// 	mockRedis.On("Get", mock.Anything, "users:1").Return(redis.NewStringResult("", redis.Nil))

// 	// Mock DB Query
// 	mockDB.PgDB.(*MockDB).On("Query", mock.Anything, mock.Anything).Return(mockRows, nil)

// 	retrievedUser, err := GetUser(mockDB, "1")
// 	assert.NoError(t, err)
// 	assert.Equal(t, user.ID, retrievedUser.ID)
// 	assert.Equal(t, "john.doe@example.com", retrievedUser.EmailAddress)

// 	mockDB.PgDB.(*MockDB).AssertExpectations(t)
// 	mockRedis.AssertExpectations(t)
// }

// func TestInsertUserInKVStore(t *testing.T) {
// 	mockRedis := &MockRedisClient{}
// 	mockDB := &database.Database{
// 		RedisDB: mockRedis,
// 	}

// 	user := &models.User{
// 		ID:           1,
// 		FirstName:    "John",
// 		LastName:     "Doe",
// 		EmailAddress: "john.doe@example.com",
// 	}

// 	var buf bytes.Buffer
// 	_ = gob.NewEncoder(&buf).Encode(user)

// 	// Mock Redis Set
// 	mockRedis.On("Set", mock.Anything, "users:1", buf.Bytes(), 2*time.Minute).Return(redis.NewStatusResult("OK", nil))

// 	err := userservice.InsertUserInKVStore(mockDB, user.ID, buf.Bytes())
// 	assert.NoError(t, err)

// 	mockRedis.AssertExpectations(t)
// }
