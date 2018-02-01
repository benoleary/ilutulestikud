package httphandler

import (
	"net/http"
)

type GetHandler interface {
	HandleGet(
		httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string)
}

type PostHandler interface {
	HandlePost(
		httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string)
}

type GetAndPostHandler interface {
	HandleGet(
		httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string)

	HandlePost(
		httpResponseWriter http.ResponseWriter, httpRequest *http.Request, relevantUriSegments []string)
}
