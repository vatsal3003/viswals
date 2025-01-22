
createmigration:
	migrate create -ext sql -dir ./migrations/ -seq ${name}

migrateup:
	migrate -path ./migrations/ -database $(POSTGRES_CONN_URL) -verbose up ${n}

migratedown:
	migrate -path ./migrations/ -database $(POSTGRES_CONN_URL) -verbose down ${n}

migrateforce:
	migrate -path ./migrations/ -database $(POSTGRES_CONN_URL) force ${n}

migrateversion:
	migrate -path ./migrations/ -database $(POSTGRES_CONN_URL) version

test:
	go test -v ./...