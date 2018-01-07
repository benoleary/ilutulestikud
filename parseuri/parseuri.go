package parseuri

import (
	"net/http"
	"strings"
)

// PathSegments returns the segments of the URI path as a slice of a string array.
func PathSegments(httpRequest *http.Request) []string {
	// The initial character is '/' so we skip it to avoid an empty string as the first element.
	return strings.Split(httpRequest.URL.Path[1:], "/")
}
