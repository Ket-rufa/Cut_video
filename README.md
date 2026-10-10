# Social Video Cutter

A web application that lets users paste social media video URLs, trim them with a timeline editor, and export MP4 clips via an async FFmpeg worker.

## Architecture

```
Frontend (Vue 3)  →  REST API (Go/Gin)  →  Redis Queue  →  Media Worker (Go + FFmpeg)
                           ↓                                        ↓
                      PostgreSQL                              Storage (local/S3)
```

## Folder Structure

```
social-video-cutter/
├── apps/
│   ├── api/           Go REST API (Gin, PostgreSQL, Redis)
│   └── web/           Vue 3 + TypeScript frontend
├── workers/
│   └── media-worker/  Go worker consuming Redis queue, running FFmpeg
├── packages/
│   └── contracts/     Shared TypeScript types
├── storage/           Local file storage (temp/source/thumbnails/exports)
├── deployments/
│   ├── docker/        Dockerfiles for api/worker/web
│   └── nginx/         Nginx reverse proxy config
├── scripts/           Helper scripts
├── docker-compose.yml
└── .env.example
```

## Requirements

- Go 1.21+
- Node.js 20+
- PostgreSQL 16+
- Redis 7+
- FFmpeg (with libx264, libopus, lavfi)
- Docker & Docker Compose (for containerized run)

## Environment Setup

```bash
cp .env.example .env
# Edit .env as needed
```

## Run with Docker (Recommended)

```bash
docker compose up --build
```

- Web UI: http://localhost
- API:    http://localhost/api/v1
- API direct: http://localhost:8080

## Run Locally

### API server
```bash
cd apps/api
go mod tidy
go run ./cmd/server
```

### Worker
```bash
cd workers/media-worker
go mod tidy
go run ./cmd/worker
```

### Frontend dev
```bash
cd apps/web
npm install
npm run dev
# opens http://localhost:5173
```

## Database Migrations

Migrations run automatically on API startup. Tables created:
- `videos` — resolved video metadata
- `video_jobs` — async job queue tracking
- `video_exports` — completed export records

## API Overview

```
POST /api/v1/videos/resolve      Resolve URL → video metadata
GET  /api/v1/videos/:id          Get video by ID
GET  /api/v1/videos              List videos
POST /api/v1/videos/:id/trim     Create trim job (queued async)
POST /api/v1/videos/:id/export   Alias for trim
GET  /api/v1/jobs/:id            Get job status & progress
GET  /api/v1/jobs/:id/progress   Progress shortcut
GET  /api/v1/jobs                List all jobs
GET  /api/v1/exports/:id         Get export record
GET  /api/v1/exports/:id/download Download exported MP4
```

## Worker Architecture

1. API creates `VideoJob` in PostgreSQL
2. API pushes `JobMessage` to Redis list `queue:jobs`
3. Worker does `BLPOP queue:jobs`
4. Worker runs FFmpeg trim → saves MP4 to `storage/exports/`
5. Worker updates job progress in DB + Redis
6. Worker saves `VideoExport` record
7. Frontend polls `GET /jobs/:id` every 2 seconds

## FFmpeg Flow

- **Original quality**: stream copy (`-c copy`) — fastest, no re-encode
- **720p/1080p/480p**: re-encode with `libx264 -preset fast`
- Test video generated via `lavfi` testsrc when no source file exists

## Frontend Flow

1. User pastes URL → `POST /videos/resolve`
2. Redirect to EditorPage
3. VideoPlayer shows stream URL
4. VideoTimeline lets user drag start/end handles
5. ExportPanel shows format/quality selects
6. Click Export → `POST /videos/:id/trim`
7. Poll job status every 2s
8. When completed → show Download button

## Security Notes

- URL validation: only http/https allowed
- SSRF protection: blocks 127.x, 192.168.x, 10.x, 172.16-31.x, 169.254.x
- FFmpeg args built from validated data — never from raw user input
- Path traversal protection in download handler
- Rate limiting: recommended to add nginx rate limiting for production

## Known Limitations

- **YouTube, TikTok, Douyin**: Hỗ trợ phân tích & tải video thực tế với yt-dlp & FFmpeg. Hỗ trợ đầy đủ link chuẩn và link rút gọn (youtu.be, vm.tiktok.com, v.douyin.com).
- **Facebook, Instagram**: Return mock/stub metadata.
- **Generic provider**: Supports direct MP4 URLs (e.g. `https://example.com/video.mp4`).
- No authentication/user accounts
- No file size validation during resolve (only during actual download)
- Worker progress uses simulation; real FFmpeg stderr parsing available in `ffmpeg/progress.go`

## Roadmap

- Real platform integration (yt-dlp adapter)
- WebSocket/SSE for real-time progress
- User authentication
- Cloud storage (S3/R2/MinIO)
- Subtitle extraction
- Watermark removal detection
- Multi-track audio
- Advanced crop/filter effects
