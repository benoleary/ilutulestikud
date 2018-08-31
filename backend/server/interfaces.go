package server

import (
	"context"
	"encoding/json"
)

type httpGetAndPostHandler interface {
	// HandleGet should return the body for the given HTTP GET request
	// along with the HTTP response code.
	HandleGet(
		requestContext context.Context,
		relevantSegments []string) (interface{}, int)

	// HandlePost should perform the relevant actions and return the body
	// for the given HTTP POST request along with the HTTP response code.
	HandlePost(
		requestContext context.Context,
		httpBodyDecoder *json.Decoder,
		relevantSegments []string) (interface{}, int)
}
