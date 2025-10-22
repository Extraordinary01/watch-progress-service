package response

import (
	"fmt"
	"net/http"
)

const (
	ContentTypeHeader = "Content-Type"
	ContentTypeValue  = "application/json"
	MessageKey        = "message"
)

func NewErrorResponse(statusCode int, msg string, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Header().Set(ContentTypeHeader, ContentTypeValue)
	_, _ = w.Write([]byte(fmt.Sprintf("{\"%s\": \"%s\"}", MessageKey, msg)))
}
