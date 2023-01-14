package v1

// Example: https://github.com/ory/hydra-consent-app-go/blob/master/main.go

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"service-account/internal/service"
	"service-account/internal/transport/http/response"
)

// signupGet godoc
// @Summary     Signup user
// @Description Get signup page
// @Tags        auth
// @Produce     html
// @Success     200 {object} object{error=string}
// @Failure     500 {object} object{error=string}
// @Router      /signup [get]
func (h *HandlerAccountManagementAPI) signupGet(context *gin.Context) {
	// Render login html.
	// TODO: csrfToken for forms.
	context.HTML(http.StatusOK, "signup.html",
		gin.H{
			"csrfToken": "",
			"action":    PathSignup,
		})
}

// signupPost godoc
// @Summary     Signup user
// @Description Signup user
// @Tags        auth
// @Produce     html
// @Success     302 {object} object{error=string}
// @Failure     400 {object} object{error=string}
// @Failure     500 {object} object{error=string}
// @Router      /signup [post]
func (h *HandlerAccountManagementAPI) signupPost(context *gin.Context) {
	submit := context.PostForm("submit")
	if submit != submitSignUp {
		response.AbortMessage(context, http.StatusBadRequest, "Unexpected submit!")
		return
	}

	// Check the user's credentials.
	var userName = context.PostForm("username")
	var userEmail = context.PostForm("email")
	var userPassword = context.PostForm("password")

	// Register user.
	inputUserData := &service.UserSignUpInput{
		Username: userName,
		Email:    userEmail,
		Password: userPassword,
	}
	if err := h.services.User.SignUp(context, inputUserData); err != nil {
		var statusCode int
		if errors.Is(err, service.ErrUserAlreadyExist) {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusInternalServerError
		}

		//// Render login html with error.
		//context.HTML(statusCode, "signup.html",
		//	gin.H{
		//		"action": PathSignup,
		//		"error":  err.Error(),
		//	},
		//)

		context.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Success redirect to main page.
	//context.Redirect(http.StatusFound, pathRoot)
	// Use http.StatusCreated or http.StatusFound
	context.JSON(http.StatusFound, gin.H{
		"redirect_url": pathRoot,
	})
}
