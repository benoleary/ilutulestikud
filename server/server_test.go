package server

// We do stick to the server package to make it easier to create server.State structs with mock handlers.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// We do not bother testing the factory method New.

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

func prepareState(statusForGet int, statusForPost int) *State {
	return &State{
		accessControlAllowedOrigin: "irrelevant",
		playerHandler:              &mockGetAndPostHandler{name: "player", getCode: statusForGet, postCode: statusForPost},
		gameHandler:                &mockGetAndPostHandler{name: "game", getCode: statusForGet, postCode: statusForPost},
	}
}

func TestState_HandleBackend(t *testing.T) {
	type args struct {
		method  string
		address string
	}

	type expected struct {
		Struct *mockReturnStruct
		Code   int
	}

	tests := []struct {
		name  string
		state *State
		args  args
		expected
	}{
		{
			name:  "getPlayer",
			state: prepareState(http.StatusOK, http.StatusOK),
			args:  args{method: http.MethodGet, address: "/backend/player/test/get/player"},
			expected: expected{
				Struct: &mockReturnStruct{Name: "player", GivenSegments: []string{"test", "get", "player"}},
				Code:   http.StatusOK,
			},
		},
		{
			name:  "postPlayer",
			state: prepareState(http.StatusOK, http.StatusOK),
			args:  args{method: http.MethodPost, address: "/backend/player/test/post/player"},
			expected: expected{
				Struct: &mockReturnStruct{Name: "player", GivenSegments: []string{"test", "post", "player"}},
				Code:   http.StatusOK,
			},
		},
		{
			name:  "getValidGame",
			state: prepareState(http.StatusOK, http.StatusOK),
			args:  args{method: http.MethodGet, address: "/backend/game/test/get/game"},
			expected: expected{
				Struct: &mockReturnStruct{Name: "game", GivenSegments: []string{"test", "get", "game"}},
				Code:   http.StatusOK,
			},
		},
		{
			name:  "postValidGame",
			state: prepareState(http.StatusOK, http.StatusOK),
			args:  args{method: http.MethodPost, address: "/backend/game/test/post/game"},
			expected: expected{
				Struct: &mockReturnStruct{Name: "game", GivenSegments: []string{"test", "post", "game"}},
				Code:   http.StatusOK,
			},
		},
		{
			name:  "getInvalidGame",
			state: prepareState(http.StatusBadRequest, http.StatusOK),
			args:  args{method: http.MethodGet, address: "/backend/game/test/get/game"},
			expected: expected{
				Struct: nil,
				Code:   http.StatusBadRequest,
			},
		},
		{
			name:  "postInvalidGame",
			state: prepareState(http.StatusOK, http.StatusBadRequest),
			args:  args{method: http.MethodPost, address: "/backend/game/test/post/game"},
			expected: expected{
				Struct: nil,
				Code:   http.StatusBadRequest,
			},
		},
		{
			name:  "getInvalidSegment",
			state: prepareState(http.StatusOK, http.StatusOK),
			args:  args{method: http.MethodGet, address: "/backend/invalid/test/get/invalid"},
			expected: expected{
				Struct: nil,
				Code:   http.StatusNotFound,
			},
		},
		{
			name:  "postInvalidSegment",
			state: prepareState(http.StatusOK, http.StatusOK),
			args:  args{method: http.MethodPost, address: "/backend/invalid/test/post/invalid"},
			expected: expected{
				Struct: nil,
				Code:   http.StatusNotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpRequest := httptest.NewRequest(tt.args.method, tt.args.address, nil)
			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			responseRecorder := httptest.NewRecorder()
			tt.state.HandleBackend(responseRecorder, httpRequest)

			if responseRecorder.Code != tt.expected.Code {
				t.Errorf(
					"%v returned wrong status %v instead of expected %v",
					tt.name,
					responseRecorder.Code,
					tt.expected.Code)
			}

			if tt.expected.Struct != nil {
				var actualStruct mockReturnStruct
				decodingError := json.NewDecoder(responseRecorder.Body).Decode(&actualStruct)

				if decodingError != nil {
					t.Fatalf(
						"%v wrote undecodable JSON: error = %v",
						tt.name,
						decodingError)
				}

				if actualStruct.Name != tt.expected.Struct.Name {
					t.Fatalf(
						"%v returned wrong struct %v instead of expected %v",
						tt.name,
						actualStruct,
						tt.expected.Struct)
				}

				for segmentIndex := 0; segmentIndex < len(tt.expected.Struct.GivenSegments); segmentIndex++ {
					if actualStruct.GivenSegments[segmentIndex] != tt.expected.Struct.GivenSegments[segmentIndex] {
						t.Fatalf(
							"%v returned wrong struct %v instead of expected %v",
							tt.name,
							actualStruct,
							tt.expected.Struct)
					}
				}
			}
		})
	}
}

// We do not bother testing the private method parsePathSegments, which is covered by HandleBackend anyway.
