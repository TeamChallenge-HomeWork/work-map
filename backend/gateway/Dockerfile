# build a base image
FROM golang:1.22.1-alpine as builder

RUN mkdir /app

WORKDIIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN CGO_ENABLE=0 go build -o gateway ./cmd/api/

RUN chmod +x /app/gateway

# build tiny image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/gateway /app

CMD ["/app/gateway"]