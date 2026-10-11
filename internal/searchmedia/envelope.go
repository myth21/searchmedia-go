package searchmedia

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const (
	Terms         = "/api/searchmedia/v1/doc/"
	jsonMediaType = "application/json; charset=UTF-8"
)

var httpStatuses = map[int]string{
	200: "OK22",
	201: "Created",
	204: "No Content",
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	410: "This method has been deprecated and is no longer available",
	429: "Too Many Requests",
	500: "Internal Server Error",
	501: "Not Implemented",
}

type MetaBody struct {
	StatusCode int    `json:"status_code"`
	StatusText string `json:"status_text"`
	Meta       any    `json:"meta"`
	Pagination any    `json:"pagination"`
	Timestamp  string `json:"timestamp"`
	RequestID  string `json:"request_id"`
	Terms      string `json:"terms"`
}

type DataBody struct {
	MetaBody
	Data any `json:"data"`
}

type ErrorBody struct {
	MetaBody
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	ID      string   `json:"id"`
	Message string   `json:"message"`
	Help    string   `json:"help"`
	Details []string `json:"details"`
}

type Pagination struct {
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	ItemsCount int `json:"items_count"`
}

// Nil meta or pagination is written as {}.
func NewMetaBody(code int, requestID string, meta, pagination any) MetaBody {
	if meta == nil {
		meta = struct{}{}
	}

	if pagination == nil {
		pagination = struct{}{}
	}

	return MetaBody{
		StatusCode: code,
		StatusText: httpStatuses[code],
		Meta:       meta,
		Pagination: pagination,
		Timestamp:  time.Now().UTC().Format("2006-01-02T15:04:05.000000-07:00"),
		RequestID:  requestID,
		Terms:      Terms,
	}
}

func NewErrorBody(code int, requestID, message, helpURL string) ErrorBody {
	return ErrorBody{
		MetaBody: NewMetaBody(code, requestID, nil, nil),
		Error:    ErrorDetails{ID: "", Message: message, Help: helpURL, Details: []string{}},
	}
}

// Leaves <, > and & unescaped and adds no trailing newline, as the other implementations do.
func writeJSON(writer http.ResponseWriter, code int, body any) {
	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(body); err != nil {
		http.Error(writer, httpStatuses[http.StatusInternalServerError], http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", jsonMediaType)
	writer.WriteHeader(code)

	_, _ = writer.Write(bytes.TrimSuffix(buffer.Bytes(), []byte("\n")))
}
