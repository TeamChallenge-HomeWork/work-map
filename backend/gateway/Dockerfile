# build a base image
FROM golang:1.22.1-alpine as builder

RUN mkdir /app

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLE=0 go build -o gateway ./cmd/api/

RUN chmod +x /app/gateway

# build tiny image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/gateway /app

COPY .env /app

EXPOSE 4001
CMD ["/app/gateway"]