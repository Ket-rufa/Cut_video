FROM golang:alpine AS builder
WORKDIR /build
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata ffmpeg curl python3 py3-pip nodejs aria2

# Install yt-dlp binary
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

# Build-time verification (Rule #21)
RUN yt-dlp --version && ffmpeg -version && ffprobe -version && node -v && aria2c --version

WORKDIR /app
COPY --from=builder /api .
EXPOSE 8080
CMD ["./api"]
