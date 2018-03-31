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
	name     string
	getCode  int
	postCode int
}

type mockReturnStruct struct {
	Name          string
	GivenSegments []string
}

func (mockHandler *mockGetAndPostHandler) HandleGet(relevantSegments []string) (interface{}, int) {
	return mockReturnStruct{Name: mockHandler.name, GivenSegments: relevantSegments[:]}, mockHandler.getCode
}

func (mockHandler *mockGetAndPostHandler) HandlePost(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int) {
	return mockReturnStruct{Name: mockHandler.name, GivenSegments: relevantSegments[:]}, mockHandler.postCode
}

func prepareState(statusForGet int, statusForPost int) *server.State {
	// It is OK to set the player handler to nil as this file just tests endpoints
	// which are not covered by requests which would get validly redirected to the
	// player endpoint handler.
	return server.New(
		"irrelevant",
		nil,
		nil,
		&mockGetAndPostHandler{name: "game", getCode: statusForGet, postCode: statusForPost})
}

// This tests that the HandleBackend function selects the correct handler and the correct function of the handler.
func TestHandleBackend(unitTest *testing.T) {
	type testArguments struct {
		method                         string
		address                        string
		bodyIsNilRatherThanEmptyObject bool
	}

	type expectedReturns struct {
		returnedStruct *mockReturnStruct
		returnedCode   int
	}

	testCases := []struct {
		name      string
		state     *server.State
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name:  "GetRoot",
			state: prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{
				method:  http.MethodGet,
				address: "/",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: nil,
				returnedCode:   http.StatusNotFound,
			},
		},
		{
			name:  "GetBackend",
			state: prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{
				method:  http.MethodGet,
				address: "/backend",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: nil,
				returnedCode:   http.StatusNotFound,
			},
		},
		{
			name:  "OptionsPlayer",
			state: prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{
				method:  http.MethodOptions,
				address: "/backend/player/test/options/player",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: nil,
				returnedCode:   http.StatusOK,
			},
		},
		{
			name:  "PutPlayer",
			state: prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{
				method:  http.MethodPut,
				address: "/backend/player/test/put/player",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: nil,
				returnedCode:   http.StatusBadRequest,
			},
		},
		{
			name:  "GetValidGame",
			state: prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{
				method:  http.MethodGet,
				address: "/backend/game/test/get/game",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: &mockReturnStruct{Name: "game", GivenSegments: []string{"test", "get", "game"}},
				returnedCode:   http.StatusOK,
			},
		},
		{
			name:  "PostValidGame",
			state: prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{
				method:  http.MethodPost,
				address: "/backend/game/test/post/game",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: &mockReturnStruct{Name: "game", GivenSegments: []string{"test", "post", "game"}},
				returnedCode:   http.StatusOK,
			},
		},
		{
			name:  "GetInvalidGame",
			state: prepareState(http.StatusBadRequest, http.StatusOK),
			arguments: testArguments{
				method:  http.MethodGet,
				address: "/backend/game/test/get/game",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: nil,
				returnedCode:   http.StatusBadRequest,
			},
		},
		{
			name:  "PostInvalidGame",
			state: prepareState(http.StatusOK, http.StatusBadRequest),
			arguments: testArguments{
				method:  http.MethodPost,
				address: "/backend/game/test/post/game",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: nil,
				returnedCode:   http.StatusBadRequest,
			},
		},
		{
			name:  "GetInvalidSegment",
			state: prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{
				method:  http.MethodGet,
				address: "/backend/invalid/test/get/invalid",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: nil,
				returnedCode:   http.StatusNotFound,
			},
		},
		{
			name:  "PostInvalidSegment",
			state: prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{
				method:  http.MethodPost,
				address: "/backend/invalid/test/post/invalid",
				bodyIsNilRatherThanEmptyObject: false,
			},
			expected: expectedReturns{
				returnedStruct: nil,
				returnedCode:   http.StatusNotFound,
			},
		},
	}

	for _, testCase := range testCases {
		unitTest.Run(testCase.name, func(unitTest *testing.T) {
			httpRequest := httptest.NewRequest(testCase.arguments.method, testCase.arguments.address, nil)
			if testCase.arguments.bodyIsNilRatherThanEmptyObject {
				httpRequest.Body = nil
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			responseRecorder := httptest.NewRecorder()
			testCase.state.HandleBackend(responseRecorder, httpRequest)

			if responseRecorder.Code != testCase.expected.returnedCode {
				unitTest.Errorf(
					"%v: returned wrong status %v instead of expected %v",
					testCase.name,
					responseRecorder.Code,
					testCase.expected.returnedCode)
			}

			if testCase.expected.returnedStruct != nil {
				var actualStruct mockReturnStruct
				decodingError := json.NewDecoder(responseRecorder.Body).Decode(&actualStruct)

				if decodingError != nil {
					unitTest.Fatalf(
						"%v: wrote undecodable JSON: error = %v",
						testCase.name,
						decodingError)
				}

				if actualStruct.Name != testCase.expected.returnedStruct.Name {
					unitTest.Fatalf(
						"%v: returned wrong struct %v instead of expected %v",
						testCase.name,
						actualStruct,
						testCase.expected.returnedStruct)
				}

				for segmentIndex := 0; segmentIndex < len(testCase.expected.returnedStruct.GivenSegments); segmentIndex++ {
					if actualStruct.GivenSegments[segmentIndex] != testCase.expected.returnedStruct.GivenSegments[segmentIndex] {
						unitTest.Fatalf(
							"%v: returned wrong struct %v instead of expected %v",
							testCase.name,
							actualStruct,
							testCase.expected.returnedStruct)
					}
				}
			}
		})
	}
}
