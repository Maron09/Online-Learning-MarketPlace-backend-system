package types

import (
	"net/http"
)











type StatusRecorder struct {
	http.ResponseWriter
	StatusCode int
}



func (rec *StatusRecorder) WriteHeader(code int) {
	rec.StatusCode = code
	rec.ResponseWriter.WriteHeader(code)
}