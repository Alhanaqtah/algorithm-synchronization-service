run:
	go run ./cmd/main.go

docs:
	swag init -g ./cmd/main.go -o ./docs