package main

import "service-account/internal/app"

// @title       Service-Account API
// @version     1.0
// @description Account management microservice.

// @contact.name  M3ik Shizuka
// @contact.url   https://m3ikshizuka.github.io/
// @contact.email m3ikshizuka@gmail.com

// @host     127.0.0.1:3000
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description OAuth 2.0 AuthZ
func main() {
	// init transport handler.
	app.Run()
}
