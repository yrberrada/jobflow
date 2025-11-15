run:
	go run ./cmd/jobflow

tidy:
	go mod tidy

build:
	go build -o jobflow ./cmd/jobflow
