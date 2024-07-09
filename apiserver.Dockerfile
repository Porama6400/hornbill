FROM golang:1.22.5-alpine as builder
WORKDIR /app
COPY . .
RUN go build -v ./cmd/apiserver/

FROM alpine:3.20.1
WORKDIR /app
RUN apk update --no-cache && apk upgrade libssl3 libcrypto3
COPY --from=builder /app/apiserver .
CMD ["/app/apiserver"]
