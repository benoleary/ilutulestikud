package server_test

// We do stick to the server package to make it easier to create server.State structs with mock handlers.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benoleary/ilutulestikud/server"
)

// This just tests that the factory method does not cause any panics, and returns a non-nil pointer.
func TestState_NewWithDefaultHandlers(testSet *testing.T) {
	type testArguments struct {
		accessControlAllowedOrigin string
	}
	testCases := []struct {
		name      string
		arguments testArguments
	}{
		{
			name: "constructorDoesNotReturnNil",
			arguments: testArguments{
				accessControlAllowedOrigin: "irrelevant",
			},
		},
	}
	for _, testCase := range testCases {
		actualState := server.NewWithDefaultHandlers(testCase.arguments.accessControlAllowedOrigin)
		if actualState == nil {
			testSet.Errorf("%v. new state was nil.", testCase.name)
		}
	}
}

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
	return server.NewWithExplicitHandlers(
		"irrelevant",
		&mockGetAndPostHandler{name: "player", getCode: statusForGet, postCode: statusForPost},
		&mockGetAndPostHandler{name: "game", getCode: statusForGet, postCode: statusForPost})
}

// This tests that the HandleBackend function selects the correct handler and the correct function of the handler.
func TestState_HandleBackend(testSet *testing.T) {
	type testArguments struct {
		method  string
		address string
	}

	type expectedReturns struct {
		returnStruct *mockReturnStruct
		returnCode   int
	}

	testCases := []struct {
		name      string
		state     *server.State
		arguments testArguments
		expected  expectedReturns
	}{
		{
			name:      "getPlayer",
			state:     prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{method: http.MethodGet, address: "/backend/player/test/get/player"},
			expected: expectedReturns{
				returnStruct: &mockReturnStruct{Name: "player", GivenSegments: []string{"test", "get", "player"}},
				returnCode:   http.StatusOK,
			},
		},
		{
			name:      "postPlayer",
			state:     prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{method: http.MethodPost, address: "/backend/player/test/post/player"},
			expected: expectedReturns{
				returnStruct: &mockReturnStruct{Name: "player", GivenSegments: []string{"test", "post", "player"}},
				returnCode:   http.StatusOK,
			},
		},
		{
			name:      "getValidGame",
			state:     prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{method: http.MethodGet, address: "/backend/game/test/get/game"},
			expected: expectedReturns{
				returnStruct: &mockReturnStruct{Name: "game", GivenSegments: []string{"test", "get", "game"}},
				returnCode:   http.StatusOK,
			},
		},
		{
			name:      "postValidGame",
			state:     prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{method: http.MethodPost, address: "/backend/game/test/post/game"},
			expected: expectedReturns{
				returnStruct: &mockReturnStruct{Name: "game", GivenSegments: []string{"test", "post", "game"}},
				returnCode:   http.StatusOK,
			},
		},
		{
			name:      "getInvalidGame",
			state:     prepareState(http.StatusBadRequest, http.StatusOK),
			arguments: testArguments{method: http.MethodGet, address: "/backend/game/test/get/game"},
			expected: expectedReturns{
				returnStruct: nil,
				returnCode:   http.StatusBadRequest,
			},
		},
		{
			name:      "postInvalidGame",
			state:     prepareState(http.StatusOK, http.StatusBadRequest),
			arguments: testArguments{method: http.MethodPost, address: "/backend/game/test/post/game"},
			expected: expectedReturns{
				returnStruct: nil,
				returnCode:   http.StatusBadRequest,
			},
		},
		{
			name:      "getInvalidSegment",
			state:     prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{method: http.MethodGet, address: "/backend/invalid/test/get/invalid"},
			expected: expectedReturns{
				returnStruct: nil,
				returnCode:   http.StatusNotFound,
			},
		},
		{
			name:      "postInvalidSegment",
			state:     prepareState(http.StatusOK, http.StatusOK),
			arguments: testArguments{method: http.MethodPost, address: "/backend/invalid/test/post/invalid"},
			expected: expectedReturns{
				returnStruct: nil,
				returnCode:   http.StatusNotFound,
			},
		},
	}

	for _, testCase := range testCases {
		testSet.Run(testCase.name, func(testSet *testing.T) {
			httpRequest := httptest.NewRequest(testCase.arguments.method, testCase.arguments.address, nil)
			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			responseRecorder := httptest.NewRecorder()
			testCase.state.HandleBackend(responseRecorder, httpRequest)

			if responseRecorder.Code != testCase.expected.returnCode {
				testSet.Errorf(
					"%v returned wrong status %v instead of expected %v",
					testCase.name,
					responseRecorder.Code,
					testCase.expected.returnCode)
			}

			if testCase.expected.returnStruct != nil {
				var actualStruct mockReturnStruct
				decodingError := json.NewDecoder(responseRecorder.Body).Decode(&actualStruct)

				if decodingError != nil {
					testSet.Fatalf(
						"%v wrote undecodable JSON: error = %v",
						testCase.name,
						decodingError)
				}

				if actualStruct.Name != testCase.expected.returnStruct.Name {
					testSet.Fatalf(
						"%v returned wrong struct %v instead of expected %v",
						testCase.name,
						actualStruct,
						testCase.expected.returnStruct)
				}

				for segmentIndex := 0; segmentIndex < len(testCase.expected.returnStruct.GivenSegments); segmentIndex++ {
					if actualStruct.GivenSegments[segmentIndex] != testCase.expected.returnStruct.GivenSegments[segmentIndex] {
						testSet.Fatalf(
							"%v returned wrong struct %v instead of expected %v",
							testCase.name,
							actualStruct,
							testCase.expected.returnStruct)
					}
				}
			}
		})
	}
}

// We do not bother testing the private method parsePathSegments, which is covered by HandleBackend anyway.
