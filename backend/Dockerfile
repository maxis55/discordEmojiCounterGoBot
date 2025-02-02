FROM golang:1.23.3-alpine AS builder
RUN mkdir /build

COPY . /build/
WORKDIR /build

RUN go get
RUN go build -o ./bin .


FROM alpine:3.20.3 AS app
COPY --from=builder /build/bin /app/
#COPY --from=builder /build/.env /app/.env

WORKDIR /app

CMD ["./bin"]