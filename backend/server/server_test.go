package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/server"
)

func prepareState() *server.State {
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
		mockPost(
			nil,
			"/backend/irrelevant",
			mockGet)

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
		testArguments argumentStruct
		expectedCode  expectedStruct
	}{
		{
			testName: "GET root",
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
			testName: "GetBackend",
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
			testName: "OptionsPlayer",
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
			testName: "PutPlayer",
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
			testName: "DeleteGame",
			testArguments: argumentStruct{
				requestMethod:                  http.MethodDelete,
				requestAddress:                 "/backend/game/test/delete/game",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedCode: expectedStruct{
				returnedCode: http.StatusBadRequest,
			},
		},
		{
			testName: "Post nil body to player",
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
			testName: "GetInvalidSegment",
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
			testName: "PostInvalidSegment",
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

			serverState := prepareState()

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			responseRecorder := httptest.NewRecorder()
			serverState.HandleBackend(responseRecorder, httpRequest)

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
