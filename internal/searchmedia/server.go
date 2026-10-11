package searchmedia

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"time"
)

const (
	SearchPath            = "/api/v1/searchmedia"
	ServiceNamesPath      = "/api/v1/searchmedia/service-names"
	ItemsLimit            = 30
	serviceTimeout        = 5 * time.Second
	requestIDHeader       = "X-Request-Id"
	// Key used to store and retrieve the request ID from the context.
	requestIDContextValue = contextKey("request_id")
)

// Custom type (like subtype for string) for context keys.
type contextKey string

var AvailableServiceNames = []string{KlipyServiceName, GoogleServiceName, GiphyServiceName}

type handler struct {
	config   Config
	client   HTTPClient
	throttle *RequestThrottle
}

func NewServer(config Config) *http.Server {
	return &http.Server{
		Addr:    config.HTTPAddr,
		Handler: NewHandler(config, NewHTTPClient(serviceTimeout), NewRequestThrottle(OneRequestIn)),

		// A header read that never finishes would otherwise hold the connection forever.
		ReadHeaderTimeout: 5 * time.Second,
	}
}

// NewHandler takes the client and the throttle as arguments, so a test can pass fakes.
func NewHandler(config Config, client HTTPClient, throttle *RequestThrottle) http.Handler {
	h := &handler{config: config, client: client, throttle: throttle}

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+ServiceNamesPath, h.serviceNames)
	mux.HandleFunc("GET "+SearchPath, h.search)

	return withRequestID(mux)
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := newRequestID()
		writer.Header().Set(requestIDHeader, requestID)

		ctx := request.Context()
		ctx = context.WithValue(
			ctx,
			requestIDContextValue,
			requestID,
		)
		request = request.WithContext(ctx)

		next.ServeHTTP(writer, request)
	})
}

func newRequestID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)

	return hex.EncodeToString(bytes)
}

func requestID(request *http.Request) string {
    value := request.Context().Value(requestIDContextValue)
    id, _ := value.(string)

    return id
}

func (h *handler) serviceNames(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, DataBody{
		MetaBody: NewMetaBody(http.StatusOK, requestID(request), nil, nil),
		Data:     AvailableServiceNames,
	})
}

func (h *handler) search(writer http.ResponseWriter, request *http.Request) {
	if err := h.throttle.ThrottleByIP(clientIP(request)); err != nil {
		h.writeError(writer, request, http.StatusTooManyRequests, httpStatuses[http.StatusTooManyRequests], h.config.DocumentationURL)
		return
	}

	query, err := ReadQuery(request.URL.Query(), AvailableServiceNames, ItemsLimit)
	if err != nil {
		h.writeError(writer, request, http.StatusBadRequest, err.Error(), h.config.SpecificationURL)
		return
	}

	items, meta := Search(h.client, h.config, query, AvailableServiceNames)

	writeJSON(writer, http.StatusOK, DataBody{
		MetaBody: NewMetaBody(http.StatusOK, requestID(request), meta, Pagination{
			Limit:      query.Limit,
			Offset:     query.Offset,
			ItemsCount: len(items),
		}),
		Data: items,
	})
}

func (h *handler) writeError(writer http.ResponseWriter, request *http.Request, code int, message, helpURL string) {
	writeJSON(writer, code, NewErrorBody(code, requestID(request), message, helpURL))
}

func clientIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}

	return host
}
