# viswals

## Overview

Data processing system using producer-consumer architecture for handling users data using Golang, PostgreSQL, Redis and RabbitMQ.

### Technologies

| Component           | Technology         |
|---------------------|--------------------|
| **Programming Language** | Golang             |
| **Database**        | PostgreSQL         |
| **Cache**           | Redis              |
| **Message Broker**  | RabbitMQ           |
| **Containerization**| Docker, Docker Compose |

### Note:
I have used  `gob` as the data interchange format instead of `JSON`. `gob` can be used as an data interchange format when your two service are written in `Go` only. `gob` is faster to serialize and deserialize than `JSON` if your two services (sender and receiver) are in `Go`. 

If later you decide to go with `JSON` so its easy to move from `gob` to `JSON`, just need to replace the encoder and decoder with marshal and unmarshal.   

## Components

### 1. CSV Ingestion
- Read data from CSV file and structure it
- Send it to RabbitMQ

### 2. CSV Processing
- Consume the messages from RabbitMQ
- Process it and insert into PostgreSQL and Redis

### 3. Users API
- REST APIs for get all users and get user by id
- Implemented SSE(Server Sent Events) handlers for sending users using SSE
- Extended from consumer service

### Flow

1. IngestCSV willl read data from CSV files
2. Publish this data to RabbitMQ
3. Consumer will consume data from RabbitMQ
4. Process it and insert into PostgreSQL and Redis
5. Expand consumer as server and expose users API 

### Required Setup
- Docker 
- Docker Compose

### APIs
| API Name              | HTTP Method | Path                | Description                              |
|-----------------------|-------------|---------------------|------------------------------------------|
| Get All Users         | GET         | `/users`            | Fetch a list of all users               |
| Get All Users         | GET         | `/users?first_name={first_name}`            | Fetch a list of all users filtered using first name               |
| Get All Users         | GET         | `/users?last_name={last_name}`            | Fetch a list of all users filtered using last name              |
| Get User by ID        | GET         | `/users/{id}`       | Fetch a single user by their ID         |
| Get All Users SSE           | GET        | `/users/sse`            | Fetch a list of all users and send to client using ServerSentEvents                       |


### Run Project

run `docker compose up --build` in root project directory (you can remove the `--build` flag after running one time)

and consumer will extended as API on `http://localhost:8080/` 


### Run Test cases

`go test -v ./...`

Execute above command to run all test cases

### Environment Variables

make `.env` file as per this example

```
ENCRYPTION_KEY      = YOUR ENCRYPTION KEY HERE
POSTGRES_CONN_URL   = YOUR POSTGRES CONNECTION URL  HERE
RABBITMQ_CONN_URL   = YOUR RABBITMQ CONNECTION URL HERE
REDIS_CONN_URL      = YOUR REDIS CONNECTION URL HERE
CONSUMER_PORT       = YOUR CONSUMER PORT  HERE
RABBITMQ_QUEUE_NAME = YOUR RABBITMQ QUEUE NAME HERE
LOG_LEVEL           = YOUR LOG LEVEL HERE
MIGRATE_DB          = true
```
    


### Project structure

- cmd
    - producer-service
        - Initialize RabbitMQ connection and start to read csv file
    - consumer-service
        - Intialize PostgreSQL, Redis, RabbitMQ connection and start to consume csv data
- internal
    - consts
        - Define constants
    - csv
        - Read csv file and send the data to RabbitMQ
        - Start consuming incoming messages from RabbitMQ
    - database
        - Initialize PostgreSQL and Redis connection
        - Close the PostgreSQL and Redis connection
        - Run migrations scripts
    - encryption
        - Encrypt the email address using AES-256 algorithm
        - Decrypt the encrypted email address using AES-256 algorithm
    - logger
        - Initialize zap logger according to development environment
    - rabbitmq
        - Publish the message to RabbitMQ
        - Consume the message from RabbitMQ
    - service
        - userservice
            - Userservice contains all database operations regarding users
    - userapi
        - Define api routes
        - Define api handlers
    - utils
        - Define all utility functions
- migrations
    - Consist all migrations scripts
- models
    - Define all models 
- web
    - Consist one HTML file to demonstrate ServerSentEvents SSE handler
