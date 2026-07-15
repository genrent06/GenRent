FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# -w -s strips debug info → smaller binary; CGO_ENABLED=0 → fully static
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o genrent ./cmd/main.go

FROM alpine:latest
# ca-certificates required for TLS (SMTP, HTTPS outbound calls)
RUN apk --no-cache add ca-certificates wget

# Run as non-root user
RUN addgroup -S genrent && adduser -S genrent -G genrent
WORKDIR /app

COPY --from=builder /app/genrent .

RUN chown -R genrent:genrent /app
USER genrent

EXPOSE 8080
CMD ["./genrent"]
