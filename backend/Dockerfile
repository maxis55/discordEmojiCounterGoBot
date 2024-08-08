FROM golang:1.22-alpine AS builder
RUN mkdir /build

ADD . /build/
WORKDIR /build

RUN go get
RUN go build -o ./bin .


FROM alpine AS app
COPY --from=builder /build/bin /app/
WORKDIR /app

CMD ["./bin"]