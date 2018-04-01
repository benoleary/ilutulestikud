package server_test

// We do stick to the server package to make it easier to create server.State structs with mock handlers.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/server"
)

// We need a struct to mock the GET and POST handlers.
type mockGetAndPostHandler struct {
	handlerName string
	getCode     int
	postCode    int
}

type mockReturnStruct struct {
	HandlerName   string
	GivenSegments []string
}

func (mockHandler *mockGetAndPostHandler) HandleGet(relevantSegments []string) (interface{}, int) {
	return mockReturnStruct{HandlerName: mockHandler.handlerName, GivenSegments: relevantSegments[:]}, mockHandler.getCode
}

func (mockHandler *mockGetAndPostHandler) HandlePost(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int) {
	return mockReturnStruct{HandlerName: mockHandler.handlerName, GivenSegments: relevantSegments[:]}, mockHandler.postCode
}

func prepareState(statusForGet int, statusForPost int) *server.State {
	// It is OK to set the player handler to nil as this file just tests endpoints
	// which are not covered by requests which would get validly redirected to the
	// player endpoint handler.
	return server.New(
		"irrelevant",
		nil,
		nil,
		&mockGetAndPostHandler{handlerName: "game", getCode: statusForGet, postCode: statusForPost})
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
		returnedStruct *mockReturnStruct
		returnedCode   int
	}

	testCases := []struct {
		testName        string
		serverState     *server.State
		testArguments   argumentStruct
		expectedReturns expectedStruct
	}{
		{
			testName:    "GET root",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodGet,
				requestAddress:                 "/",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedReturns: expectedStruct{
				returnedStruct: nil,
				returnedCode:   http.StatusNotFound,
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
			expectedReturns: expectedStruct{
				returnedStruct: nil,
				returnedCode:   http.StatusNotFound,
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
			expectedReturns: expectedStruct{
				returnedStruct: nil,
				returnedCode:   http.StatusOK,
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
			expectedReturns: expectedStruct{
				returnedStruct: nil,
				returnedCode:   http.StatusBadRequest,
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
			expectedReturns: expectedStruct{
				returnedStruct: nil,
				returnedCode:   http.StatusBadRequest,
			},
		},
		{
			testName:    "GetValidGame",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodGet,
				requestAddress:                 "/backend/game/test/get/game",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedReturns: expectedStruct{
				returnedStruct: &mockReturnStruct{HandlerName: "game", GivenSegments: []string{"test", "get", "game"}},
				returnedCode:   http.StatusOK,
			},
		},
		{
			testName:    "PostValidGame",
			serverState: prepareState(http.StatusOK, http.StatusOK),
			testArguments: argumentStruct{
				requestMethod:                  http.MethodPost,
				requestAddress:                 "/backend/game/test/post/game",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expectedReturns: expectedStruct{
				returnedStruct: &mockReturnStruct{HandlerName: "game", GivenSegments: []string{"test", "post", "game"}},
				returnedCode:   http.StatusOK,
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
			expectedReturns: expectedStruct{
				returnedStruct: nil,
				returnedCode:   http.StatusBadRequest,
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
			expectedReturns: expectedStruct{
				returnedStruct: nil,
				returnedCode:   http.StatusBadRequest,
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
			expectedReturns: expectedStruct{
				returnedStruct: nil,
				returnedCode:   http.StatusNotFound,
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
			expectedReturns: expectedStruct{
				returnedStruct: nil,
				returnedCode:   http.StatusNotFound,
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

			if responseRecorder.Code != testCase.expectedReturns.returnedCode {
				unitTest.Errorf(
					"%v: returned wrong status %v instead of expected %v",
					testCase.testName,
					responseRecorder.Code,
					testCase.expectedReturns.returnedCode)
			}

			if testCase.expectedReturns.returnedStruct != nil {
				var actualStruct mockReturnStruct
				decodingError := json.NewDecoder(responseRecorder.Body).Decode(&actualStruct)

				if decodingError != nil {
					unitTest.Fatalf(
						"%v: wrote undecodable JSON: error = %v",
						testCase.testName,
						decodingError)
				}

				if actualStruct.HandlerName != testCase.expectedReturns.returnedStruct.HandlerName {
					unitTest.Fatalf(
						"%v: returned wrong struct %v instead of expected %v",
						testCase.testName,
						actualStruct,
						testCase.expectedReturns.returnedStruct)
				}

				for segmentIndex := 0; segmentIndex < len(testCase.expectedReturns.returnedStruct.GivenSegments); segmentIndex++ {
					if actualStruct.GivenSegments[segmentIndex] != testCase.expectedReturns.returnedStruct.GivenSegments[segmentIndex] {
						unitTest.Fatalf(
							"%v: returned wrong struct %v instead of expected %v",
							testCase.testName,
							actualStruct,
							testCase.expectedReturns.returnedStruct)
					}
				}
			}
		})
	}
}
