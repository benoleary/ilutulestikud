package server

import (
	"encoding/json"
)

type httpGetAndPostHandler interface {
	HandleGet(relevantSegments []string) (interface{}, int)

	HandlePost(httpBodyDecoder *json.Decoder, relevantSegments []string) (interface{}, int)
}
