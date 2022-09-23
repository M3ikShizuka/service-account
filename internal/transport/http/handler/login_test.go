package handler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"golang.org/x/net/context"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"service-account/internal/domain"
	"service-account/internal/service"
	mock_service "service-account/internal/service/mocks"
	"testing"
)

type mockBehavior func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string)

type TestTable struct {
	name               string
	challenge          string
	mockBehavior       mockBehavior
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

func (t *TestTable) GetMockBehavior() mockBehavior {
	return t.mockBehavior
}

func setWorkDir() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func initArrange(ctrl *gomock.Controller, testCase *TestTable) *Handler {
	mockOAuth2 := mock_service.NewMockOAuth2(ctrl)
	testCase.GetMockBehavior()(mockOAuth2, nil, testCase.GetChallenge())

	services := &service.Services{
		OAuth2: mockOAuth2,
	}

	return NewHandler(services)
}

func initEndpoint() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.LoadHTMLGlob("./web/template/*")
	return r
}

func TestHandlerAPI_loginGet(t *testing.T) {
	setWorkDir()

	testTable := []TestTableLoginGet{
		{
			TestTable: TestTable{
				name:      "OK, NOT authorized already",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					mockOAuth.EXPECT().GetLoginRequest(gomock.Any(), challenge).Return(&domain.OA2LoginRequest{
						Skip:    false,
						Subject: "foo@bar.com",
						Hint:    "",
					}, nil)
				},
				expectedStatusCode: 200,
			},
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
		},
		{
			TestTable: TestTable{
				name:      "OK, authorized already",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					subject := "foo@bar.com"

					mockOAuth.EXPECT().GetLoginRequest(gomock.Any(), challenge).Return(&domain.OA2LoginRequest{
						Skip:    true,
						Subject: subject,
						Hint:    "",
					}, nil)

					mockOAuth.EXPECT().AcceptLoginRequest(gomock.Any(), challenge, subject, true, int64(3600)).Return("redirectToURL", nil)
				},
				expectedStatusCode: 302,
			},
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
		},
		{
			TestTable: TestTable{
				name:      "BAD, login_challenge not set",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
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
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					mockOAuth.EXPECT().GetLoginRequest(gomock.Any(), challenge).Return(nil, errors.New("Test error"))
				},
				expectedStatusCode: 500,
			},
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
		},
		{
			TestTable: TestTable{
				name:      "BAD, authorized already but AcceptLoginRequest error",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					subject := "foo@bar.com"

					mockOAuth.EXPECT().GetLoginRequest(gomock.Any(), challenge).Return(&domain.OA2LoginRequest{
						Skip:    true,
						Subject: subject,
						Hint:    "",
					}, nil)

					mockOAuth.EXPECT().AcceptLoginRequest(gomock.Any(), challenge, subject, true, int64(3600)).Return("", errors.New("Test error"))
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
			handlerAPI := initArrange(ctrl, &testCase.TestTable)

			// Init Endpoint
			r := initEndpoint()
			requestURL := pathLogin
			r.GET(requestURL, handlerAPI.loginGet)

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

func TestHandlerAPI_loginPost(t *testing.T) {
	setWorkDir()

	testTable := []TestTableLoginPost{
		{
			TestTable: TestTable{
				name: "BAD, challenge is not set",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					// Nothing
				},
				expectedStatusCode: 400,
			},
		},
		{
			TestTable: TestTable{
				name:      "BAD, submit is unknown",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
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
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					mockOAuth.EXPECT().
						RejectLoginRequest(gomock.Any(), challenge, "access_denied", "The resource owner denied the request").
						Return("redirectTo", nil)
				},
				expectedStatusCode: 302,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&submit=" + submitDenyAccess,
		},
		{
			TestTable: TestTable{
				name:      "OK, reject login request",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					mockOAuth.EXPECT().
						RejectLoginRequest(gomock.Any(), challenge, "access_denied", "The resource owner denied the request").
						Return("", errors.New("Test error"))
				},
				expectedStatusCode: 500,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&submit=" + submitDenyAccess,
		},
		{
			TestTable: TestTable{
				name:      "OK, AuthN is bad",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					// Nothing
				},
				expectedStatusCode: 200,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&submit=" + submitLogIn,
		},
		{
			TestTable: TestTable{
				name:      "BAD, get login request",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					mockOAuth.EXPECT().
						GetLoginRequest(gomock.Any(), challenge).
						Return(nil, errors.New("Test error"))
				},
				expectedStatusCode: 500,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&submit=" + submitLogIn,
		},
		{
			TestTable: TestTable{
				name:      "BAD, accept login request",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					mockOAuth.EXPECT().
						GetLoginRequest(gomock.Any(), challenge).
						Return(nil, nil)
					mockOAuth.EXPECT().
						AcceptLoginRequest(gomock.Any(), challenge, "foo@bar.com", false, int64(3600)).
						Return("", errors.New("Test error"))
				},
				expectedStatusCode: 500,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&submit=" + submitLogIn,
		},
		{
			TestTable: TestTable{
				name:      "OK, accept login request",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					mockOAuth.EXPECT().
						GetLoginRequest(gomock.Any(), challenge).
						Return(nil, nil)
					mockOAuth.EXPECT().
						AcceptLoginRequest(gomock.Any(), challenge, "foo@bar.com", false, int64(3600)).
						Return("redirectTo", nil)
				},
				expectedStatusCode: 302,
			},
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&submit=" + submitLogIn,
		},
		{
			TestTable: TestTable{
				name:      "OK, accept login request and remember",
				challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
				mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
					mockOAuth.EXPECT().
						GetLoginRequest(gomock.Any(), challenge).
						Return(nil, nil)
					mockOAuth.EXPECT().
						AcceptLoginRequest(gomock.Any(), challenge, "foo@bar.com", true, int64(3600)).
						Return("redirectTo", nil)
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
			handlerAPI := initArrange(ctrl, &testCase.TestTable)

			// Init Endpoint
			r := initEndpoint()
			requestURL := pathLogin
			r.POST(requestURL, handlerAPI.loginPost)

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
