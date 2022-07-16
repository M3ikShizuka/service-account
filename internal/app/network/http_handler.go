package network

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

const (
	pathUser string = "/user"
)

func InitHandlers() {
	httpHost := os.Getenv("HTTP_HOST")
	if httpHost == "" {
		httpHost = "0.0.0.0"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	hostAddr := httpHost + ":" + httpPort

	router := gin.Default()
	router.NoRoute(handlerNotFound)
	router.GET(pathUser, handlerUserGet)
	router.POST(pathUser, handlerUserPost)
	err := router.Run(hostAddr)
	log.Fatal(err)
}

func handlerNotFound(pContext *gin.Context) {
	pContext.Writer.WriteHeader(http.StatusNotFound)
	pContext.Writer.Write([]byte("Page not found! dontknown！"))
	return
}

func handlerUserGet(pContext *gin.Context) {
	// !!! Call busines logic func() which return data. !!!
	data := User{
		Id:   31,
		Name: "M3ik Shizuka",
	}
	pContext.IndentedJSON(http.StatusOK, data)
}

func handlerUserPost(pContext *gin.Context) {
	var dataUser User

	if err := pContext.BindJSON(&dataUser); err != nil {
		return
	}

	pContext.IndentedJSON(http.StatusCreated, dataUser)
}
