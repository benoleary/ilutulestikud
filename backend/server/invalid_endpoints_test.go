package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/server"
	"github.com/benoleary/ilutulestikud/backend/server/endpoint/parsing"
)

var mockContextProvider = &server.BackgroundContextProvider{}

type mockEndpointHandler struct {
	TestReference    *testing.T
	TestErrorForGet  error
	TestErrorForPost error
	ReturnInterface  interface{}
	ReturnCode       int
}

func ErrorEndpointHandler(unitTest *testing.T) *mockEndpointHandler {
	return &mockEndpointHandler{
		TestReference:    unitTest,
		TestErrorForGet:  fmt.Errorf("GET not intended"),
		TestErrorForPost: fmt.Errorf("POST not intended"),
		ReturnInterface:  nil,
		ReturnCode:       -1,
	}
}

// HandleGet parses an HTTP GET request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (mockHandler *mockEndpointHandler) HandleGet(
	requestContext context.Context,
	relevantSegments []string) (interface{}, int) {
	if mockHandler.TestErrorForGet != nil {
		mockHandler.TestReference.Fatalf(
			"HandleGet(%v) called: %v",
			relevantSegments,
			mockHandler.TestErrorForGet)
	}

	return mockHandler.ReturnInterface, mockHandler.ReturnCode
}

// HandlePost parses an HTTP POST request and responds with the appropriate function.
// This implements part of github.com/benoleary/ilutulestikud/server.httpGetAndPostHandler.
func (mockHandler *mockEndpointHandler) HandlePost(
	requestContext context.Context,
	httpBodyDecoder *json.Decoder,
	relevantSegments []string) (interface{}, int) {
	if mockHandler.TestErrorForPost != nil {
		mockHandler.TestReference.Fatalf(
			"HandlePost(%v, %v) called: %v",
			httpBodyDecoder,
			relevantSegments,
			mockHandler.TestErrorForPost)
	}

	return mockHandler.ReturnInterface, mockHandler.ReturnCode
}

func TestMockPostReturnsErrorForMalformedBody(unitTest *testing.T) {
	// The json encoder is quite robust, but one way to trigger an error is to try
	// to encode a function pointer.
	_, decodingError :=
		mockPost(
			nil,
			"/backend/irrelevant",
			mockGet)

	if decodingError == nil {
		unitTest.Fatal("No decoding error")
	}
}

func TestRejectInvalidRequestsBeforeCallingHandler(unitTest *testing.T) {
	testCases := []struct {
		testName                       string
		requestMethod                  string
		requestAddress                 string
		bodyIsNilRatherThanEmptyObject bool
		expectedCode                   int
	}{
		{
			testName:                       "GET root",
			requestMethod:                  http.MethodGet,
			requestAddress:                 "/",
			bodyIsNilRatherThanEmptyObject: false,
			expectedCode:                   http.StatusNotFound,
		},
		{
			testName:                       "GET backend",
			requestMethod:                  http.MethodGet,
			requestAddress:                 "/backend",
			bodyIsNilRatherThanEmptyObject: false,
			expectedCode:                   http.StatusNotFound,
		},
		{
			testName:                       "OPTIONS player",
			requestMethod:                  http.MethodOptions,
			requestAddress:                 "/backend/player/test/options/player",
			bodyIsNilRatherThanEmptyObject: false,
			expectedCode:                   http.StatusOK,
		},
		{
			testName:                       "PUT player",
			requestMethod:                  http.MethodPut,
			requestAddress:                 "/backend/player/test/put/player",
			bodyIsNilRatherThanEmptyObject: false,
			expectedCode:                   http.StatusBadRequest,
		},
		{
			testName:                       "DELETE game",
			requestMethod:                  http.MethodDelete,
			requestAddress:                 "/backend/game/test/delete/game",
			bodyIsNilRatherThanEmptyObject: false,
			expectedCode:                   http.StatusBadRequest,
		},
		{
			testName:                       "POST nil body to player",
			requestMethod:                  http.MethodPost,
			requestAddress:                 "/backend/player/test/post/nil",
			bodyIsNilRatherThanEmptyObject: true,
			expectedCode:                   http.StatusBadRequest,
		},
		{
			testName:                       "GET invalid segment",
			requestMethod:                  http.MethodGet,
			requestAddress:                 "/backend/invalid/test/get/invalid",
			bodyIsNilRatherThanEmptyObject: false,
			expectedCode:                   http.StatusNotFound,
		},
		{
			testName:                       "POST invalid segment",
			requestMethod:                  http.MethodPost,
			requestAddress:                 "/backend/invalid/test/post/invalid",
			bodyIsNilRatherThanEmptyObject: false,
			expectedCode:                   http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			httpRequest :=
				httptest.NewRequest(
					testCase.requestMethod,
					testCase.requestAddress,
					nil)
			if testCase.bodyIsNilRatherThanEmptyObject {
				httpRequest.Body = nil
			}

			// It is OK to set the player and game handlers to nil as this file just tests
			// endpoints which are not covered by requests which would get validly redirected
			// to either of the endpoint handlers.
			serverState :=
				server.New(mockContextProvider, "irrelevant to tests", nil, nil, nil)

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			responseRecorder := httptest.NewRecorder()
			serverState.HandleBackend(responseRecorder, httpRequest)

			if responseRecorder.Code != testCase.expectedCode {
				unitTest.Errorf(
					"%v: returned wrong status %v instead of expected %v",
					testCase.testName,
					responseRecorder.Code,
					testCase.expectedCode)
			}
		})
	}
}

func TestSelectCorrectHandlerForValidRequest(unitTest *testing.T) {
	wrongHandler := ErrorEndpointHandler(unitTest)
	getHandler := ErrorEndpointHandler(unitTest)
	getHandler.TestErrorForGet = nil
	getHandler.ReturnInterface = "success"
	getHandler.ReturnCode = http.StatusOK
	postHandler := ErrorEndpointHandler(unitTest)
	postHandler.TestErrorForPost = nil
	postHandler.ReturnInterface = "success"
	postHandler.ReturnCode = http.StatusOK

	testCases := []struct {
		testName       string
		requestMethod  string
		requestAddress string
		playerHandler  *mockEndpointHandler
		gameHandler    *mockEndpointHandler
	}{
		{
			testName:       "GET player",
			requestMethod:  http.MethodGet,
			requestAddress: "/backend/player",
			playerHandler:  getHandler,
			gameHandler:    wrongHandler,
		},
		{
			testName:       "GET game",
			requestMethod:  http.MethodGet,
			requestAddress: "/backend/game",
			playerHandler:  wrongHandler,
			gameHandler:    getHandler,
		},
		{
			testName:       "POST player",
			requestMethod:  http.MethodPost,
			requestAddress: "/backend/player",
			playerHandler:  postHandler,
			gameHandler:    wrongHandler,
		},
		{
			testName:       "POST game",
			requestMethod:  http.MethodPost,
			requestAddress: "/backend/game",
			playerHandler:  wrongHandler,
			gameHandler:    postHandler,
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			httpRequest :=
				httptest.NewRequest(
					testCase.requestMethod,
					testCase.requestAddress,
					nil)

			serverState :=
				server.NewWithGivenHandlers(
					mockContextProvider,
					"irrelevant to tests",
					nil,
					testCase.playerHandler,
					testCase.gameHandler)

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			responseRecorder := httptest.NewRecorder()
			serverState.HandleBackend(responseRecorder, httpRequest)

			if responseRecorder.Code != http.StatusOK {
				unitTest.Errorf(
					"%v: returned wrong status %v instead of expected %v",
					testCase.testName,
					responseRecorder.Code,
					http.StatusOK)
			}
		})
	}
}

func TestWrapReturnedError(unitTest *testing.T) {
	testHandler := ErrorEndpointHandler(unitTest)
	testHandler.TestErrorForGet = nil
	expectedError := fmt.Errorf("expected error")
	testHandler.ReturnInterface = expectedError
	testHandler.ReturnCode = http.StatusBadRequest

	httpRequest :=
		httptest.NewRequest(
			http.MethodGet,
			"/backend/player",
			nil)

	serverState :=
		server.NewWithGivenHandlers(
			mockContextProvider,
			"irrelevant to tests",
			nil,
			testHandler,
			nil)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	responseRecorder := httptest.NewRecorder()
	serverState.HandleBackend(responseRecorder, httpRequest)

	if responseRecorder.Code != testHandler.ReturnCode {
		unitTest.Errorf(
			"GET returned wrong status %v instead of expected %v",
			responseRecorder.Code,
			testHandler.ReturnCode)
	}

	var errorForBody parsing.ErrorForBody
	errorFromUnmarshall := json.Unmarshal(responseRecorder.Body.Bytes(), &errorForBody)
	if errorFromUnmarshall != nil {
		unitTest.Errorf(
			"error when unmarshalling JSON: %v",
			errorFromUnmarshall)
	}

	if errorForBody.Error != expectedError.Error() {
		unitTest.Errorf(
			"response body %v did not have expected Error %v",
			testHandler.ReturnInterface,
			expectedError)
	}
}
