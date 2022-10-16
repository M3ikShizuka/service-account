package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// TODO: Refactor.

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

const (
	pathUser string = "/user"
)

func userGet(context *gin.Context) {
	// !!! Call busines logic func() which return data. !!!
	data := User{
		Id:   31,
		Name: "M3ik Shizuka",
	}
	context.IndentedJSON(http.StatusOK, data)
}

func userPost(context *gin.Context) {
	var dataUser User
	if err := context.BindJSON(&dataUser); err != nil {
		return
	}
	context.IndentedJSON(http.StatusCreated, dataUser)
}
