
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

# .DEFAULT_GOAL := build