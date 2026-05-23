FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/insider-league ./cmd/api

FROM alpine:3.20

RUN adduser -D -H appuser
USER appuser

COPY --from=builder /bin/insider-league /bin/insider-league

EXPOSE 8080

ENTRYPOINT ["/bin/insider-league"]
