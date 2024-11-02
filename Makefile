.PHONY:
.SILENT:
.DEFAULT_GOAL := run
include .env
export $(shell sed 's/=.*//' .env)


build:
	go mod download && CGO_ENABLED=0 GOOS=linux go build -o ./.bin/app ./cmd/app/main.go
run:
	go run ./cmd/app/main.go
swag:
	swag init -g internal/app/app.go