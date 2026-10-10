package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type Config struct {
	Env        string
	Server     ServerConfig
	DB         DBConfig
	Redis      RedisConfig
	Storage    StorageConfig
	Downloader DownloaderConfig
	Limits     LimitsConfig
	CORS       CORSConfig
	AI         AIConfig
}

type AIConfig struct {
	GeminiAPIKey string
	Model        string
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
}

type RedisConfig struct {
	Host string
	Port string
}

type StorageConfig struct {
	Driver string
	Root   string
}

type DownloaderConfig struct {
	YTDLPPath   string
	FFmpegPath  string
	FFprobePath string
}

type LimitsConfig struct {
	MaxDurationSeconds   int
	MaxSizeMB            int
	ExportExpireHours    int
	AnalyzeRateLimit     int // requests per minute
	CreateJobRateLimit   int // requests per minute
	MaxConcurrentGlobal  int
	MaxConcurrentPerUser int
	TempCleanupHours     int
}

type CORSConfig struct {
	Origins string
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		d.Host, d.Port, d.DBName, d.User, d.Password)
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func resolveExecutable(preferredEnv, exeName string, extraPaths ...string) string {
	if envVal := os.Getenv(preferredEnv); envVal != "" {
		if _, err := os.Stat(envVal); err == nil {
			return envVal
		}
	}

	if p, err := exec.LookPath(exeName); err == nil {
		return p
	}

	for _, p := range extraPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Default fallback
	if os.Getenv(preferredEnv) != "" {
		return os.Getenv(preferredEnv)
	}
	return exeName
}

func Load() *Config {
	localAppData := os.Getenv("LOCALAPPDATA")
	wingetYTDLP := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages", "yt-dlp.yt-dlp_Microsoft.Winget.Source_8wekyb3d8bbwe", "yt-dlp.exe")
	wingetFFmpeg := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages", "yt-dlp.FFmpeg_Microsoft.Winget.Source_8wekyb3d8bbwe", "ffmpeg-N-125875-g5d4d3bdc61-win64-gpl", "bin", "ffmpeg.exe")
	wingetFFprobe := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages", "yt-dlp.FFmpeg_Microsoft.Winget.Source_8wekyb3d8bbwe", "ffmpeg-N-125875-g5d4d3bdc61-win64-gpl", "bin", "ffprobe.exe")

	ytdlpResolved := resolveExecutable("YTDLP_PATH", "yt-dlp", "/usr/bin/yt-dlp", "/usr/local/bin/yt-dlp", wingetYTDLP)
	ffmpegResolved := resolveExecutable("FFMPEG_PATH", "ffmpeg", "/usr/bin/ffmpeg", "/usr/local/bin/ffmpeg", wingetFFmpeg)
	ffprobeResolved := resolveExecutable("FFPROBE_PATH", "ffprobe", "/usr/bin/ffprobe", "/usr/local/bin/ffprobe", wingetFFprobe)

	return &Config{
		Env: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port: getEnv("API_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			DBName:   getEnv("POSTGRES_DB", "video_cutter"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
		},
		Storage: StorageConfig{
			Driver: getEnv("STORAGE_DRIVER", "local"),
			Root:   getEnv("STORAGE_ROOT", "./storage"),
		},
		Downloader: DownloaderConfig{
			YTDLPPath:   ytdlpResolved,
			FFmpegPath:  ffmpegResolved,
			FFprobePath: ffprobeResolved,
		},
		Limits: LimitsConfig{
			MaxDurationSeconds:   getEnvInt("MAX_VIDEO_DURATION_SECONDS", 86400), // 24 hours
			MaxSizeMB:            getEnvInt("MAX_VIDEO_SIZE_MB", 51200),          // 50 GB
			ExportExpireHours:    getEnvInt("EXPORT_EXPIRE_HOURS", 24),
			AnalyzeRateLimit:     getEnvInt("VIDEO_ANALYZE_RATE_LIMIT", 60),     // 60 req/min
			CreateJobRateLimit:   getEnvInt("VIDEO_CREATE_JOB_RATE_LIMIT", 20),  // 20 req/min
			MaxConcurrentGlobal:  getEnvInt("VIDEO_MAX_CONCURRENT_GLOBAL", 4),   // 4 concurrent downloads globally
			MaxConcurrentPerUser: getEnvInt("VIDEO_MAX_CONCURRENT_PER_USER", 2), // max 2 downloads per user
			TempCleanupHours:     getEnvInt("TEMP_CLEANUP_HOURS", 6),            // clean temp files older than 6h
		},
		CORS: CORSConfig{
			Origins: getEnv("CORS_ORIGINS", "http://localhost:5173"),
		},
		AI: AIConfig{
			GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
			Model:        getEnv("GEMINI_MODEL", "gemini-3.5-flash-lite"),
		},
	}
}
