# Service Account
User's account management service.

## Deployment
Deploy to Kubernetes with [k8s-microservices](https://github.com/M3ikShizuka/k8s-microservices).

## Configurations
Change params in [configs/config.yml](https://github.com/M3ikShizuka/service-account/blob/develop/configs/config.yml) file if necessary to run the service locally (usually for develop and debug).  
Use environment variables to set parameters when you deploy the service in a kubernetes cluster.

## Unit tests
```bash
make test-unit
```

## Generate Swagger API (OpenAPI) from code 
Install swag
```bash
make swagger-install
```

Generate API
```bash
make swagger-generate-api-doc
```

Open site: `yourdomain`/swagger/index.html

## Libs
* [Argon2id](https://pkg.go.dev/golang.org/x/crypto/argon2) - user account password hash algorithm for store in database.
* [gorm](https://gorm.io) - ORM library.
* [viper](https://github.com/spf13/viper) - configuration. 
* [zap](https://github.com/uber-go/zap) - logging.
* [gomock](https://github.com/golang/mock) - mocking framework.
* [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) - Sql mock driver for golang to test database interactions.
* [swag](https://github.com/swaggo/swag) - Automatically generate RESTful API documentation with Swagger 2.0 for Go.