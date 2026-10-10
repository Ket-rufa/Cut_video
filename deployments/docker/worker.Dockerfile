FROM golang:alpine AS builder
WORKDIR /build
COPY workers/media-worker/go.mod workers/media-worker/go.sum ./
RUN go mod download
COPY workers/media-worker/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata ffmpeg curl python3 nodejs aria2

# Install yt-dlp binary
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

# Build-time verification
RUN yt-dlp --version && ffmpeg -version && ffprobe -version && node -v && aria2c --version

WORKDIR /app
COPY --from=builder /worker .
CMD ["./worker"]
