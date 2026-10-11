package searchmedia

import "os"

// DefaultHTTPAddr is the default address used by the HTTP server.
// The app listens on port 80 inside the container. The host port is configured separately in Docker Compose.
const DefaultHTTPAddr = ":80"

// DefaultSpecificationURL is the default URL of the Swagger UI.
// DefaultDocumentationURL is the default URL of the API documentation.
const (
	DefaultSpecificationURL = "/api/searchmedia/v1/swagger/"
	DefaultDocumentationURL = "/api/searchmedia/v1/doc/"
)

// Config holds the application configuration.
type Config struct {
	HTTPAddr         string
	SpecificationURL string
	DocumentationURL string

	// API credentials for external search services.
	GiphyKey  string
	KlipyKey  string
	GoogleKey string

	// Google Custom Search Engine ID.
	GoogleCX string
}

// ConfigFromEnv loads the application configuration from environment variables.
func ConfigFromEnv() Config {
	return Config{
		HTTPAddr:         envOr("SEARCHMEDIA_HTTP_ADDR", DefaultHTTPAddr),
		SpecificationURL: envOr("SEARCHMEDIA_SPECIFICATION_URL", DefaultSpecificationURL),
		DocumentationURL: envOr("SEARCHMEDIA_DOCUMENTATION_URL", DefaultDocumentationURL),
		GiphyKey:         os.Getenv("SEARCHMEDIA_GIPHY_KEY"),
		KlipyKey:         os.Getenv("SEARCHMEDIA_KLIPY_KEY"),
		GoogleKey:        os.Getenv("SEARCHMEDIA_GOOGLE_KEY"),
		GoogleCX:         os.Getenv("SEARCHMEDIA_GOOGLE_CX"),
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}