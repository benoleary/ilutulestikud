package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/server"
)

func prepareState(statusForGet int, statusForPost int) *server.State {
	// It is OK to set the player and game handlers to nil as this file just tests
	// endpoints which are not covered by requests which would get validly redirected
	// to either of the endpoint handlers.
	return server.New(
		"irrelevant",
		nil,
		nil,
		nil)
}

func TestMockPostReturnsErrorForMalformedBody(unitTest *testing.T) {
	// The json encoder is quite robust, but one way to trigger an error is to try
	// to encode a function pointer.
	_, decodingError :=
		server.MockPost(
			nil,
			"/backend/irrelevant",
			server.MockGet)

	if decodingError == nil {
		unitTest.Fatal("No decoding error")
	}
}

// This tests that the HandleBackend function selects the correct handler and the correct function of the handler.
func TestHandleBackend(unitTest *testing.T) {
	type argumentStruct struct {
		requestMethod                  string
		requestAddress                 string
		bodyIsNilRatherThanEmptyObject bool
	}

	type expectedStruct struct {
		returnedCode int
	}

	testCases := []struct {
		testName      string
		serverState   *server.State
		testArguments argumentStruct
		expectedCode  expectedStruct
	}{
		{
			testName:    "GET root",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodGet,
				requestAddress:                 "/",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusNotFound,
			},
		},
		{
			testName:    "GetBackend",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodGet,
				requestAddress:                 "/backend",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusNotFound,
			},
		},
		{
			testName:    "OptionsPlayer",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodOptions,
				requestAddress:                 "/backend/player/test/options/player",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusOK,
			},
		},
		{
			testName:    "PutPlayer",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodPut,
				requestAddress:                 "/backend/player/test/put/player",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusBadRequest,
			},
		},
		{
			testName:    "Post nil body to player",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodPost,
				requestAddress:                 "/backend/player/test/post/nil",
				bodyIsNilRatherThanEmptyObject: true,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusBadRequest,
			},
		},
		{
			testName:    "GetInvalidGame",
			serverState: prepareState(http.StatusBadRequest, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodGet,
				requestAddress:                 "/backend/game/test/get/game",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusBadRequest,
			},
		},
		{
			testName:    "PostInvalidGame",
			serverState: prepareState(http.StatusOK, http.StatusBadRequest),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodPost,
				requestAddress:                 "/backend/game/test/post/game",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusBadRequest,
			},
		},
		{
			testName:    "GetInvalidSegment",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodGet,
				requestAddress:                 "/backend/invalid/test/get/invalid",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusNotFound,
			},
		},
		{
			testName:    "PostInvalidSegment",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodPost,
				requestAddress:                 "/backend/invalid/test/post/invalid",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusNotFound,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.testName, func(unitTest *testing.T) {
			httpRequest := httptest.NewRequest(testCase.testArguments.requestMethod, testCase.testArguments.requestAddress, nil)
			if testCase.testArguments.bodyIsNilRatherThanEmptyObject {
				httpRequest.Body = nil
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			responseRecorder := httptest.NewRecorder()
			testCase.serverState.HandleBackend(responseRecorder, httpRequest)

			if responseRecorder.Code != testCase.expectedCode.returnedCode {
				unitTest.Errorf(
					"%v: returned wrong status %v instead of expected %v",
					testCase.testName,
					responseRecorder.Code,
					testCase.expectedCode.returnedCode)
			}
		})
	}
}
