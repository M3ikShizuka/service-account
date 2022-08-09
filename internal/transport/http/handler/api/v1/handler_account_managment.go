package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

const (
	pathUser string = "/user"
)

func (handler *HandlerAPIv1) initHandlersAccountManagment(router *gin.RouterGroup) {
	router.GET(pathUser, handlerUserGet)
	router.POST(pathUser, handlerUserPost)
}

func handlerUserGet(context *gin.Context) {
	// !!! Call busines logic func() which return data. !!!
	data := User{
		Id:   31,
		Name: "M3ik Shizuka",
	}
	context.IndentedJSON(http.StatusOK, data)
}

func handlerUserPost(context *gin.Context) {
	var dataUser User
	if err := context.BindJSON(&dataUser); err != nil {
		return
	}
	context.IndentedJSON(http.StatusCreated, dataUser)
}
