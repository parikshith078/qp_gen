FROM golang:1.24-alpine as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o brokerApp ./cmd/api

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/brokerApp .

CMD [ "/app/brokerApp" ]
