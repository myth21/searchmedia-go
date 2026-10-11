# Search Media API

A media search service written in **Go**. It searches GIFs and images across Klipy, Giphy, and Google Custom Search simultaneously and returns a unified JSON envelope response.

## Endpoints

| Route | What it does |
|---|---|
| `GET /api/v1/searchmedia` | Search. At most one request per 5 seconds from one IP |
| `GET /api/v1/searchmedia/service-names` | List of available services |

### Search parameters

| Parameter | Required | Description |
|---|---|---|
| `text` | yes | Search text, 1 to 32 characters |
| `service` | no | Service: `klipy`, `google`, or `giphy`. Can be repeated (`service=klipy&service=giphy`). If omitted, all services are queried |
| `limit` | no | How many items to return (1 to 30). Defaults to 30 |
| `offset` | no | Pagination offset. Defaults to 0 |

Invalid parameters return `400`, too frequent requests return `429`. If one of the services fails, results from the remaining services are still returned.

Example:

```bash
curl 'http://127.0.0.1:8083/api/v1/searchmedia?text=cats&service=giphy&limit=5'
```

---

## Structure

```text
.
├── cmd/
│   └── searchmedia/
│       └── main.go              # Application entry point
├── go.mod                       # Go module definition
└── internal/
    └── searchmedia/             # Private application logic
        ├── client.go            # HTTP client for external APIs
        ├── config.go            # Application configuration & env parsing
        ├── envelope.go          # Unified JSON response envelope models
        ├── giphy.go             # Giphy API client & response handling
        ├── google.go            # Google Custom Search client & handling
        ├── item.go              # Core media item domains/models
        ├── klipy.go             # Klipy API client & handling
        ├── mapping.go           # Data transformation helpers
        ├── plan.go              # Request limit distribution logic across services
        ├── query.go             # Query string parsing and validation
        ├── server.go            # HTTP server, routing, and endpoints
        ├── service.go           # Core business logic orchestrating API calls
        ├── throttle.go          # Per-IP rate limiting middleware
        ├── *_test.go            # Unit and integration tests
        └── testdata/            # Test fixtures
            ├── edge-cases/      # Mock responses with missing/null fields
            └── recorded/        # Real recorded API responses for mapping tests
```

---

## Configuration

All settings are configured using environment variables prefixed with `SEARCHMEDIA_`.

| Variable | Default | Description |
|---|---|---|
| `SEARCHMEDIA_HOST` | `0.0.0.0` | IP address the server listens on |
| `SEARCHMEDIA_PORT` | `8080` | Server HTTP port |
| `SEARCHMEDIA_GIPHY_KEY` | | Giphy API key |
| `SEARCHMEDIA_KLIPY_KEY` | | Klipy API key |
| `SEARCHMEDIA_GOOGLE_KEY` | | Google Custom Search API key |
| `SEARCHMEDIA_GOOGLE_CX` | | Google Search Engine ID (cx) |
| `SEARCHMEDIA_SPECIFICATION_URL` | | OpenAPI/Swagger specification link in error responses |
| `SEARCHMEDIA_DOCUMENTATION_URL` | | Documentation link in `429` error responses |
| `SEARCHMEDIA_HOST` | `127.0.0.1` | Host address used by `compose.yaml` |
| `SEARCHMEDIA_HOST_PORT` | `8083` | Host port used by `compose.yaml` |

> The application will start even without API keys. Services lacking valid keys will return an error status and be marked as unavailable in the response envelope.

---

## Running the Application

### Running locally

**Requirements:** Go 1.21+

1. Run directly:
   ```bash
   go run ./cmd/searchmedia
   ```

2. Or load environment variables from `.env` file first:
   ```bash
   export $(cat .env | xargs) && go run ./cmd/searchmedia
   ```

3. Build binary executable:
   ```bash
   go build -o bin/searchmedia ./cmd/searchmedia
   ./bin/searchmedia
   ```

### Running with Docker

1. Copy the example environment file and fill in your keys:
   ```bash
   cp .env.example .env
   ```

2. Build and start the container:
   ```bash
   docker compose up -d --build
   ```

3. View application logs:
   ```bash
   docker compose logs -f
   ```

4. Stop the service:
   ```bash
   docker compose down
   ```

---

## Testing

Tests utilize standard Go testing tools and fixture data stored in `internal/searchmedia/testdata/`. Tests do not issue real network requests.

Run all tests:
```bash
go test ./...
```

Run tests with detailed output (`-v`) and race condition detection (`-race`):
```bash
go test -v -race ./...
```

Run tests for a specific file or package:
```bash
go test -v ./internal/searchmedia -run TestQuery
```

---

## Known Limitations

- Services are queried sequentially with individual timeouts. A single search request can take up to several seconds if external APIs respond slowly.
- Rate limit counters (throttling) are maintained in-memory per instance. If deployed behind a load balancer with multiple replicas, per-IP rate limits will not be synchronized across instances.