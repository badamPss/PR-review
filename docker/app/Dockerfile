FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pr-review ./cmd/pr-review

FROM alpine:latest
RUN apk --no-cache add ca-certificates postgresql-client
WORKDIR /root/

COPY --from=builder /app/pr-review .
COPY --from=builder /app/config ./config
COPY --from=builder /app/db/migration ./db/migration
COPY --from=builder /app/scripts ./scripts

EXPOSE 8080
CMD ["./pr-review"]


