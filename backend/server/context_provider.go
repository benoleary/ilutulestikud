package server

import (
	"context"
	"net/http"
)

// ContextProvider is a means to provide context.Context objects from
// the HTTP request without having to know about things such as how
// Google App Engine has its own static function for providing one.
type ContextProvider interface {
	// FromRequest should provide a context from the request.
	FromRequest(httpRequest *http.Request) context.Context
}

// BackgroundContextProvider simply provides the background context,
// ignoring the given HTTP request.
type BackgroundContextProvider struct {
}

// FromRequest simply provides the background context, ignoring the
// given HTTP request.
func (backgroundContextProvider *BackgroundContextProvider) FromRequest(
	httpRequest *http.Request) context.Context {
	return context.Background()
}
