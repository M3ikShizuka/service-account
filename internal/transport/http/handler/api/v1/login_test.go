package v1

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

const pathAPI = "/api/v1"

func setWorkDir() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestHandlerAPIv1_loginGet(t *testing.T) {
	setWorkDir()

	type mockBehavior func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string)

	testTable := []struct {
		name               string
		inputBody          string
		challenge          string
		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name:      "OK, NOT authorized already",
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
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
		{
			name:      "OK, authorized already",
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
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
		{
			name:      "BAD, login_challenge not set",
			inputBody: "",
			challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
			mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
				// Nothing
			},
			expectedStatusCode: 400,
		},
		{
			name:      "BAD, GetLoginRequest error",
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
			challenge: "2f5d20b9e8f0404aafe01978a8d92a45",
			mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
				mockOAuth.EXPECT().GetLoginRequest(gomock.Any(), challenge).Return(nil, errors.New("Test error"))
			},
			expectedStatusCode: 500,
		},
		{
			name:      "BAD, authorized already but AcceptLoginRequest error",
			inputBody: "login_challenge=2f5d20b9e8f0404aafe01978a8d92a45",
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
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//// Arrange
			mockOAuth2 := mock_service.NewMockOAuth2(ctrl)
			testCase.mockBehavior(mockOAuth2, nil, testCase.challenge)

			services := &service.Services{
				OAuth2: mockOAuth2,
			}
			handlerAPIv1 := HandlerAPIv1{services}

			// Init Endpoint
			gin.SetMode(gin.ReleaseMode)
			r := gin.New()
			r.LoadHTMLGlob("./web/template/*")
			requestURL := fmt.Sprintf("%s%s", pathAPI, pathLogin)
			r.GET(requestURL, handlerAPIv1.loginGet)

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

func TestHandlerAPIv1_loginPost(t *testing.T) {
	setWorkDir()

	type mockBehavior func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string)

	testTable := []struct {
		name               string
		requestGetParams   string
		requestBody        string
		challenge          string
		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name: "BAD, challenge is not set",
			mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
				// Nothing
			},
			expectedStatusCode: 400,
		},
		{
			name:        "BAD, submit is unknown",
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45",
			challenge:   "2f5d20b9e8f0404aafe01978a8d92a45",
			mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
				// Nothing
			},
			expectedStatusCode: 400,
		},
		{
			name:        "BAD, submit is Deny Access",
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&submit=" + submitDenyAccess,
			challenge:   "2f5d20b9e8f0404aafe01978a8d92a45",
			mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
				mockOAuth.EXPECT().
					RejectLoginRequest(gomock.Any(), challenge, "access_denied", "The resource owner denied the request").
					Return("redirectTo", nil)
			},
			expectedStatusCode: 302,
		},
		{
			name:        "OK, reject login request",
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&submit=" + submitDenyAccess,
			challenge:   "2f5d20b9e8f0404aafe01978a8d92a45",
			mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
				mockOAuth.EXPECT().
					RejectLoginRequest(gomock.Any(), challenge, "access_denied", "The resource owner denied the request").
					Return("", errors.New("Test error"))
			},
			expectedStatusCode: 500,
		},
		{
			name:        "OK, AuthN is bad",
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&submit=" + submitLogIn,
			challenge:   "2f5d20b9e8f0404aafe01978a8d92a45",
			mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
				// Nothing
			},
			expectedStatusCode: 200,
		},
		{
			name:        "BAD, get login request",
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&submit=" + submitLogIn,
			challenge:   "2f5d20b9e8f0404aafe01978a8d92a45",
			mockBehavior: func(mockOAuth *mock_service.MockOAuth2, ctx *context.Context, challenge string) {
				mockOAuth.EXPECT().
					GetLoginRequest(gomock.Any(), challenge).
					Return(nil, errors.New("Test error"))
			},
			expectedStatusCode: 500,
		},
		{
			name:        "BAD, accept login request",
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&submit=" + submitLogIn,
			challenge:   "2f5d20b9e8f0404aafe01978a8d92a45",
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
		{
			name:        "OK, accept login request",
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&submit=" + submitLogIn,
			challenge:   "2f5d20b9e8f0404aafe01978a8d92a45",
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
		{
			name:        "OK, accept login request and remember",
			requestBody: "challenge=2f5d20b9e8f0404aafe01978a8d92a45&email=foo%40bar.com&password=foobar&remember=true&submit=" + submitLogIn,
			challenge:   "2f5d20b9e8f0404aafe01978a8d92a45",
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
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//// Arrange
			mockOAuth2 := mock_service.NewMockOAuth2(ctrl)
			testCase.mockBehavior(mockOAuth2, nil, testCase.challenge)

			services := &service.Services{
				OAuth2: mockOAuth2,
			}
			handlerAPIv1 := HandlerAPIv1{services}

			// Init Endpoint
			gin.SetMode(gin.ReleaseMode)
			r := gin.New()
			r.LoadHTMLGlob("./web/template/*")
			requestURL := fmt.Sprintf("%s%s", pathAPI, pathLogin)
			r.POST(requestURL, handlerAPIv1.loginPost)

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
