FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o url-shortener main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/url-shortener /app/url-shortener
COPY --from=builder /app/static /app/static
COPY --from=builder /app/configs /app/configs

EXPOSE 8000 2121

ENTRYPOINT ["/app/url-shortener"]
