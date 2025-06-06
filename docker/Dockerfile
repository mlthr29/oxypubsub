FROM golang:1.24.2-alpine AS builder

COPY . /app
WORKDIR /app

RUN go test ./tests/... && go build -o oxypubsub main.go

FROM alpine:latest

COPY --from=builder /app/oxypubsub /oxypubsub

EXPOSE 8080 8081

CMD ["/oxypubsub"]