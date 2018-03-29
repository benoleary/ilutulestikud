package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

const breaksBase64 = "\\/\\\\\\?" // This should unescape to \/\\\? in the tests.

// MockGet creates a mock HTTP GET request and sends it to the given
// server.State and returns an object containing the recorded response.
func MockGet(
	testState *State,
	mockAddress string) *httptest.ResponseRecorder {
	httpRequest := httptest.NewRequest(http.MethodGet, mockAddress, nil)

	return mockHandleBackend(testState, httpRequest)
}

// MockPost creates a mock HTTP POST request, encoding the given object
// into a JSON body for the request, and sends it to the given server.State
// and returns an object containing the recorded response.
func MockPost(
	testState *State,
	mockAddress string,
	jsonBody interface{}) (*httptest.ResponseRecorder, error) {
	bytesBuffer := new(bytes.Buffer)
	encodingError := json.NewEncoder(bytesBuffer).Encode(jsonBody)

	if encodingError != nil {
		return nil, encodingError
	}

	httpRequest :=
		httptest.NewRequest(
			http.MethodPost,
			mockAddress,
			bytesBuffer)

	return mockHandleBackend(testState, httpRequest), nil
}

// MockPostWithDirectBody creates a mock HTTP POST request, encoding the
// given string directly as the body of the request, and sends it to the
// given server.State and returns an object containing the recorded response.
func MockPostWithDirectBody(
	testState *State,
	mockAddress string,
	jsonBody string) *httptest.ResponseRecorder {
	httpRequest :=
		httptest.NewRequest(
			http.MethodPost,
			mockAddress,
			strings.NewReader(jsonBody))

	return mockHandleBackend(testState, httpRequest)
}

func mockHandleBackend(
	testState *State,
	httpRequest *http.Request) *httptest.ResponseRecorder {
	// We create a ResponseRecorder (which satisfies http.ResponseWriter)
	// to record the response.
	responseRecorder := httptest.NewRecorder()

	testState.HandleBackend(responseRecorder, httpRequest)

	return responseRecorder
}
