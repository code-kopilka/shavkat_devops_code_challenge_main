ARG GOLANG_VERSION=1.22.2

FROM --platform=$TARGETPLATFORM golang:${GOLANG_VERSION} AS build

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# mattn/go-sqlite3 requires CGO
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=1 go build -o app main.go

RUN chmod +x app

# base includes CGO bindings
FROM --platform=$TARGETPLATFORM gcr.io/distroless/base AS runtime

# Create non-root user (distroless base image has user 65532)
USER 65532:65532

# Copy only the binary (not .env file - use environment variables instead)
COPY --from=build /build/app /app

# Note: Healthcheck requires a tool like curl/wget which distroless doesn't have
# Options:
# 1. Use a different base image (e.g., alpine) that includes curl
# 2. Add a minimal HTTP client binary to the distroless image
# 3. Use external health checking (e.g., ALB health checks, monitoring tools)
# The application has a /health endpoint available at http://localhost:PORT/health
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ["/healthcheck-binary", "http://localhost:3000/health"] || exit 1

ENTRYPOINT ["/app"]
