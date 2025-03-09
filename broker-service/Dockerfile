FROM golang:1.24-alpine AS builder

# Build-time arguments
ARG BUILD_VERSION="1.0.0"
ARG BUILD_DATE
ARG COMMIT_SHA

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o brokerApp ./cmd/api

FROM alpine:latest

# Add labels to the final image
ARG BUILD_VERSION="1.0.0"
ARG BUILD_DATE
ARG COMMIT_SHA

LABEL org.opencontainers.image.version="${BUILD_VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${COMMIT_SHA}"
LABEL org.opencontainers.image.title="QP Broker Service"
LABEL org.opencontainers.image.description="Broker service for QP Generator"

WORKDIR /app

COPY --from=builder /app/brokerApp .
COPY --from=builder /app/migrations ./migrations

CMD [ "/app/brokerApp" ]



