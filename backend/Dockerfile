FROM golang:1.23.3-alpine AS builder
WORKDIR /build

# Cache deps in a separate layer so changes to .go files don't redownload.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin .


FROM alpine:3.20.3 AS app
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /build/bin /app/bin
USER app

CMD ["./bin"]
