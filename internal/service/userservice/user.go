package userservice

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/vatsal3003/viswals/internal/database"
	"github.com/vatsal3003/viswals/internal/encryption"
	"github.com/vatsal3003/viswals/models"
)

func InsertUser(db *database.Database, user *models.User) error {
	_, err := db.PgDB.Exec("INSERT INTO users VALUES ($1, $2, $3, $4, $5, $6, $7, $8);", user.ID, user.FirstName, user.LastName, user.EmailAddress, user.CreatedAt, user.DeletedAt, user.MergedAt, user.ParentUserID)
	if err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Code == "23505" && pqErr.Constraint == "users_pkey" {
			return nil
		}
		return err
	}

	return nil
}

func InsertUserInKVStore(db *database.Database, userID int, user []byte) error {
	status := db.RedisDB.Set(context.Background(), "users:"+strconv.Itoa(userID), user, 2*time.Minute)
	if status.Err() != nil {
		// If there is error during inserting in cache, do nothing as its not critical task
		log.Println("ERROR failed to set the user:" + status.Err().Error())
	}

	return nil
}

func GetAllUsers(db *database.Database, filters map[string]string) ([]*models.User, error) {
	whereClause := " WHERE "

	hasAnyFilterApplied := false

	fname, ok := filters["first_name"]
	if ok {
		if hasAnyFilterApplied {
			whereClause += "OR first_name ILIKE '" + fname + "%' "
		} else {
			whereClause += "first_name ILIKE '" + fname + "%' "
		}
		hasAnyFilterApplied = true
	}

	lname, ok := filters["last_name"]
	if ok {
		if hasAnyFilterApplied {
			whereClause += "OR last_name ILIKE '" + lname + "%' "
		} else {
			whereClause += "last_name ILIKE '" + lname + "%' "
		}
		hasAnyFilterApplied = true
	}

	var users []*models.User

	query := `SELECT id, first_name, last_name, email_address, created_at, deleted_at, merged_at, parent_user_id FROM users`

	if len(filters) != 0 {
		query = query + whereClause
	}

	rows, err := db.PgDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := &models.User{}
		err = rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.EmailAddress, &user.CreatedAt, &user.DeletedAt, &user.MergedAt, &user.ParentUserID)
		if err != nil {
			return nil, err
		}

		user.EmailAddress, err = encryption.Decrypt(user.EmailAddress)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func GetUser(db *database.Database, userID string) (*models.User, error) {
	var user = new(models.User)

	res, err := db.RedisDB.Get(context.Background(), "users:"+userID).Bytes()
	if err == redis.Nil {
		row, err := db.PgDB.Query("SELECT id, first_name, last_name, email_address, created_at, deleted_at, merged_at, parent_user_id FROM users WHERE id = $1", userID)
		if err != nil {
			return nil, err
		}
		defer row.Close()

		row.Next()
		err = row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.EmailAddress, &user.CreatedAt, &user.DeletedAt, &user.MergedAt, &user.ParentUserID)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		err = gob.NewEncoder(&buf).Encode(user)
		if err != nil {
			return nil, err
		}

		_ = db.RedisDB.Set(context.Background(), "users:"+strconv.Itoa(user.ID), buf.Bytes(), 2*time.Minute).Err()

		user.EmailAddress, err = encryption.Decrypt(user.EmailAddress)
		if err != nil {
			return nil, err
		}

		return user, nil
	} else if err != nil {
		return nil, err
	} else {

		var user models.User
		err := gob.NewDecoder(bytes.NewReader(res)).Decode(&user)
		if err != nil {
			return nil, err
		}

		user.EmailAddress, err = encryption.Decrypt(user.EmailAddress)
		if err != nil {
			return nil, err
		}

		return &user, nil
	}
}
