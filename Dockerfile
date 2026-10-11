FROM golang:1.24 AS builder

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/searchmedia ./cmd/searchmedia

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/searchmedia ./searchmedia

EXPOSE 80

ENTRYPOINT ["/app/searchmedia"]