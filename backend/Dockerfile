FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /scheduler ./cmd/scheduler
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates curl

COPY --from=builder /api /app/api
COPY --from=builder /scheduler /app/scheduler
COPY --from=builder /worker /app/worker
COPY migrations /app/migrations

EXPOSE 8080

CMD ["/app/api"]