FROM golang:1.26-alpine AS build
ARG VERSION="dev"

# Set the working directory
WORKDIR /build

# Install git
RUN apk add --no-cache git

# Copy module files first so dependency download is cached as its own layer
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X cfg.version=${VERSION} " \
    -o /bin/render-mcp-server main.go

# Make a stage to run the app
FROM gcr.io/distroless/base-debian12
# Set the working directory
WORKDIR /server
# Copy the binary from the build stage
COPY --from=build /bin/render-mcp-server .
# Set default config path (inside container)
ENV RENDER_CONFIG_PATH=/config/mcp-server.yaml
# Use ENTRYPOINT instead of CMD so that additional user-provided args are passed to the server.
# Default to HTTP transport so Railway serves /mcp; override args if needed.
ENTRYPOINT ["./render-mcp-server"]
CMD ["--transport", "http"]
