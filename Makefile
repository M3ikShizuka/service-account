swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/swaggo/files@latest
	go install github.com/swaggo/gin-swagger@latest

swagger-generate-api-doc:
	swag init -g cmd/service-account/main.go --output api

test-unit:
	go test ./...