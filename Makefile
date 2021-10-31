
# compile and create binary
.PHONY: build
build:
	go build -o service ./cmd/service/main.go

# set DB-variables and start the application
.PHONY: start
start:
	go build -o service ./cmd/service/main.go && \
	export DB_HOST="localhost" && \
	export DB_PORT="5432" && \
	export DB_USER="gotuber" && \
	export DB_PASSWORD="Passw0rd" && \
	export DB_NAME="go_tube" && \
	~/pg-n-go/Serg_Kotovsky_5/hw5/service

# start test functions
.PHONY: test
test:
	go test ./... -count 1

# start integration tests
.PHONY: int
int:
	go test ./... -tags=integration_tests -v -count 1

# .DEFAULT_GOAL := build