package v1

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"service-account/internal/domain"
	"service-account/internal/service"
	mock_service "service-account/internal/service/mocks"
	"testing"
	"time"
)

type mockBehaviorOAuth2 func(mockOAuth *mock_service.MockOAuth2, challenge string)
type mockBehaviorUser func(mockUser *mock_service.MockUser)

type TestTable struct {
	name               string
	challenge          string
	userData           interface{}
	mockBehaviorOAuth2 mockBehaviorOAuth2
	mockBehaviorUser   mockBehaviorUser
	expectedStatusCode int
}

type TestTableLoginGet struct {
	TestTable
	inputBody string
}

type TestTableLoginPost struct {
	TestTable
	requestGetParams string
	requestBody      string
}

func (t *TestTable) GetChallenge() string {
	return t.challenge
}

func (t *TestTable) GetUserData() interface{} {
	return t.userData
}

func (t *TestTable) GetmockBehaviorOAuth2() mockBehaviorOAuth2 {
	return t.mockBehaviorOAuth2
}

func (t *TestTable) GetmockBehaviorUser() mockBehaviorUser {
	return t.mockBehaviorUser
}

func setWorkDir() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func initArrange(ctrl *gomock.Controller, testCase *TestTable) *HandlerAccountManagementAPI {
	mockOAuth2 := mock_service.NewMockOAuth2(ctrl)
	testCase.GetmockBehaviorOAuth2()(mockOAuth2, testCase.GetChallenge())

	mockUser := mock_service.NewMockUser(ctrl)
	testCase.GetmockBehaviorUser()(mockUser)

	services := service.NewService(
		nil,
		nil,
		mockOAuth2,
		mockUser,
	)

	return NewHandlerAccountManagementAPI(services)
}

func initEndpoint() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.LoadHTMLGlob("./web/template/*")
	return r
}

func TestHandlerAccountManagementAPI_loginGet(t *testing.T) {
	setWorkDir()

	testTable := []TestTableLoginGet{
		{
			TestTable: TestTable{
				name:      "OK, NOT authorized already",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				userData:  &service.UserSignInInput{},
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					mockOAuth.EXPECT().GetLoginRequest(gomock.Any(), challenge).Return(&domain.OA2LoginRequest{
						Skip:    false,
						Subject: "foo@bar.com",
						Hint:    "",
					}, nil)
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					// Nothing
				},
				expectedStatusCode: 200,
			},
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
		},
		{
			TestTable: TestTable{
				name:      "OK, authorized already",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					subject := "foo@bar.com"

					mockOAuth.EXPECT().GetLoginRequest(gomock.Any(), challenge).Return(&domain.OA2LoginRequest{
						Skip:    true,
						Subject: subject,
						Hint:    "",
					}, nil)

					mockOAuth.EXPECT().AcceptLoginRequest(gomock.Any(), challenge, subject, true, int64(3600)).Return("redirectToURL", nil)
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					// Nothing
				},
				expectedStatusCode: 302,
			},
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
		},
		{
			TestTable: TestTable{
				name:      "BAD, login_challenge not set",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					// Nothing
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					// Nothing
				},
				expectedStatusCode: 400,
			},
			inputBody: "",
		},
		{
			TestTable: TestTable{
				name:      "BAD, GetLoginRequest error",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					mockOAuth.EXPECT().GetLoginRequest(gomock.Any(), challenge).Return(nil, errors.New("Test error"))
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					// Nothing
				},
				expectedStatusCode: 500,
			},
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
		},
		{
			TestTable: TestTable{
				name:      "BAD, authorized already but AcceptLoginRequest error",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					subject := "foo@bar.com"

					mockOAuth.EXPECT().GetLoginRequest(gomock.Any(), challenge).Return(&domain.OA2LoginRequest{
						Skip:    true,
						Subject: subject,
						Hint:    "",
					}, nil)

					mockOAuth.EXPECT().AcceptLoginRequest(gomock.Any(), challenge, subject, true, int64(3600)).Return("", errors.New("Test error"))
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					// Nothing
				},
				expectedStatusCode: 500,
			},
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//// Arrange
			HandlerAccountManagementAPI := initArrange(ctrl, &testCase.TestTable)

			// Init Endpoint
			r := initEndpoint()
			requestURL := pathSignin
			r.GET(requestURL, HandlerAccountManagementAPI.signinGet)

			// Create Request
			w := httptest.NewRecorder()
			requestURLWithParams := fmt.Sprintf("%s?%s", requestURL, testCase.inputBody)
			req := httptest.NewRequest("GET", requestURLWithParams, nil)

			//// Act
			// Make Request
			r.ServeHTTP(w, req)

			//// Assert
			assert.Equal(t, w.Code, testCase.expectedStatusCode)
		})
	}
}

func TestHandlerAccountManagementAPI_loginPost(t *testing.T) {
	setWorkDir()

	testPassHash, _ := hex.DecodeString("92b2723f184a5f9b17ba52b88079391b") // pass: 1234567890qwerty
	testUser := &domain.User{
		Id:               1,
		Username:         "test",
		Email:            "foo@bar.com",
		PasswordHash:     testPassHash,
		DateRegistration: time.Now(),
		DateLastOnline:   time.Now(),
	}

	testTable := []TestTableLoginPost{
		{
			TestTable: TestTable{
				name: "BAD, challenge is not set",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					// Nothing
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					// Nothing
				},
				expectedStatusCode: 400,
			},
		},
		{
			TestTable: TestTable{
				name:      "BAD, submit is unknown",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					// Nothing
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					// Nothing
				},
				expectedStatusCode: 400,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45",
		},
		{
			TestTable: TestTable{
				name:      "BAD, submit is Deny Access",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					mockOAuth.EXPECT().
						RejectLoginRequest(gomock.Any(), challenge, "access_denied", "The resource owner denied the request").
						Return("redirectTo", nil)
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					// Nothing
				},
				expectedStatusCode: 302,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&submit=" + submitDenyAccess,
		},
		{
			TestTable: TestTable{
				name:      "OK, reject login request",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					mockOAuth.EXPECT().
						RejectLoginRequest(gomock.Any(), challenge, "access_denied", "The resource owner denied the request").
						Return("", service.ErrUserNotFound)
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					// Nothing
				},
				expectedStatusCode: 500,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&submit=" + submitDenyAccess,
		},
		{
			TestTable: TestTable{
				name:      "OK, AuthN is bad",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					// Nothing
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					mockUser.EXPECT().
						SignIn(gomock.Any(), &service.UserSignInInput{}).
						Return(nil, service.ErrUserNotFound)
				},
				expectedStatusCode: 400,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&submit=" + submitLogIn,
		},
		{
			TestTable: TestTable{
				name:      "BAD, get login request",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					mockOAuth.EXPECT().
						GetLoginRequest(gomock.Any(), challenge).
						Return(nil, errors.New("Test error"))
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					mockUser.EXPECT().
						SignIn(gomock.Any(), &service.UserSignInInput{"foo@bar.com", "foobar"}).
						Return(testUser, nil)
				},
				expectedStatusCode: 500,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&submit=" + submitLogIn,
		},
		{
			TestTable: TestTable{
				name:      "BAD, accept login request",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					mockOAuth.EXPECT().
						GetLoginRequest(gomock.Any(), challenge).
						Return(nil, nil)
					mockOAuth.EXPECT().
						AcceptLoginRequest(gomock.Any(), challenge, "1", false, int64(3600)).
						Return("", errors.New("Test error"))
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					mockUser.EXPECT().
						SignIn(gomock.Any(), &service.UserSignInInput{"foo@bar.com", "foobar"}).
						Return(testUser, nil)
				},
				expectedStatusCode: 500,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&submit=" + submitLogIn,
		},
		{
			TestTable: TestTable{
				name:      "OK, accept login request",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					mockOAuth.EXPECT().
						GetLoginRequest(gomock.Any(), challenge).
						Return(nil, nil)
					mockOAuth.EXPECT().
						AcceptLoginRequest(gomock.Any(), challenge, "1", false, int64(3600)).
						Return("redirectTo", nil)
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					mockUser.EXPECT().
						SignIn(gomock.Any(), &service.UserSignInInput{"foo@bar.com", "foobar"}).
						Return(testUser, nil)
				},
				expectedStatusCode: 302,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&submit=" + submitLogIn,
		},
		{
			TestTable: TestTable{
				name:      "OK, accept login request and remember",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehaviorOAuth2: func(mockOAuth *mock_service.MockOAuth2, challenge string) {
					mockOAuth.EXPECT().
						GetLoginRequest(gomock.Any(), challenge).
						Return(nil, nil)
					mockOAuth.EXPECT().
						AcceptLoginRequest(gomock.Any(), challenge, "1", true, int64(3600)).
						Return("redirectTo", nil)
				},
				mockBehaviorUser: func(mockUser *mock_service.MockUser) {
					mockUser.EXPECT().
						SignIn(gomock.Any(), &service.UserSignInInput{"foo@bar.com", "foobar"}).
						Return(testUser, nil)
				},
				expectedStatusCode: 302,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&remember=true&submit=" + submitLogIn,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//// Arrange
			HandlerAccountManagementAPI := initArrange(ctrl, &testCase.TestTable)

			// Init Endpoint
			r := initEndpoint()
			requestURL := pathSignin
			r.POST(requestURL, HandlerAccountManagementAPI.signinPost)

			// Create Request
			w := httptest.NewRecorder()
			requestURLWithParams := fmt.Sprintf("%s?%s", requestURL, testCase.requestGetParams)
			req := httptest.NewRequest("POST", requestURLWithParams, bytes.NewBufferString(testCase.requestBody))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			//// Act
			// Make Request
			r.ServeHTTP(w, req)

			//// Assert
			assert.Equal(t, w.Code, testCase.expectedStatusCode)
		})
	}
}
