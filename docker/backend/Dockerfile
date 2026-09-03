FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" \
    -o /out/weavelens ./cmd/weavelens

FROM alpine:3.21 AS runtime

WORKDIR /app
RUN mkdir -p /app/data && chown -R 65534:65534 /app
COPY --from=builder --chown=65534:65534 /out/weavelens /app/weavelens
USER 65534:65534

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ["wget", "-q", "-O", "-", "http://localhost:8080/health"]

ENTRYPOINT ["/app/weavelens"]
