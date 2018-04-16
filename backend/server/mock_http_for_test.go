package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benoleary/ilutulestikud/backend/server"
)

// segmentTranslatorForTest returns the standard base-32 translator.
func segmentTranslatorForTest() server.EndpointSegmentTranslator {
	return &server.Base32Translator{}
}

// mockGet creates a mock HTTP GET request and sends it to the given
// server.State and returns an object containing the recorded response.
func mockGet(
	testState *server.State,
	mockAddress string) *httptest.ResponseRecorder {
	httpRequest := httptest.NewRequest(http.MethodGet, mockAddress, nil)

	return mockHandleBackend(testState, httpRequest)
}

// mockPost creates a mock HTTP POST request, encoding the given object
// into a JSON body for the request, and sends it to the given server.State
// and returns an object containing the recorded response.
func mockPost(
	testState *server.State,
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

// mockPostWithDirectBody creates a mock HTTP POST request, encoding the
// given string directly as the body of the request, and sends it to the
// given server.State and returns an object containing the recorded response.
func mockPostWithDirectBody(
	testState *server.State,
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
	testState *server.State,
	httpRequest *http.Request) *httptest.ResponseRecorder {
	// We create a ResponseRecorder (which satisfies http.ResponseWriter)
	// to record the response.
	responseRecorder := httptest.NewRecorder()

	testState.HandleBackend(responseRecorder, httpRequest)

	return responseRecorder
}

func assertResponseIsCorrect(
	unitTest *testing.T,
	testIdentifier string,
	responseRecorder *httptest.ResponseRecorder,
	encodingError error,
	expectedCode int) {
	if encodingError != nil {
		unitTest.Fatalf(
			testIdentifier+"/encoding error: %v",
			encodingError)
	}

	if responseRecorder == nil {
		unitTest.Fatalf(testIdentifier + "/endpoint returned nil response.")
	}

	if responseRecorder.Code != expectedCode {
		unitTest.Fatalf(
			testIdentifier+"/did not return expected HTTP code %v, instead was %v.",
			expectedCode,
			responseRecorder.Code)
	}
}
