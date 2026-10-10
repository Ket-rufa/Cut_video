package downloader

import (
	"context"
	"net"
	"testing"
	"social-video-cutter/api/internal/storage"
)

func TestSecurity_IsAllowedHost(t *testing.T) {
	tests := []struct {
		host    string
		allowed bool
	}{
		{"youtube.com", true},
		{"www.youtube.com", true},
		{"m.youtube.com", true},
		{"youtu.be", true},
		{"youtube.com.attacker.com", false},
		{"notyoutube.com", false},
		{"localhost", false},
		{"tiktok.com", false}, // not in youtube allowlist
	}

	for _, tt := range tests {
		got := IsAllowedHost(tt.host, allowedYouTubeHosts)
		if got != tt.allowed {
			t.Errorf("IsAllowedHost(%q, youtube) = %v; want %v", tt.host, got, tt.allowed)
		}
	}
}

func TestSecurity_IsBlockedIP(t *testing.T) {
	tests := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"::1", true},
		{"0.0.0.0", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true}, // AWS/GCP metadata
		{"8.8.8.8", false},
		{"1.1.1.1", false},
	}

	for _, tt := range tests {
		parsed := net.ParseIP(tt.ip)
		got := IsBlockedIP(parsed)
		if got != tt.blocked {
			t.Errorf("IsBlockedIP(%q) = %v; want %v", tt.ip, got, tt.blocked)
		}
	}
}

func TestSecurity_ValidateYouTubeURL(t *testing.T) {
	valid := []string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ",
		"https://youtube.com/shorts/abcdef12345",
	}

	for _, u := range valid {
		_, err := ValidateYouTubeURL(u)
		if err != nil {
			t.Errorf("ValidateYouTubeURL(%q) unexpected error: %v", u, err)
		}
	}

	invalid := []string{
		"https://youtube.com.attacker.com/watch?v=fake",
		"http://localhost/watch?v=fake",
		"http://127.0.0.1:8080",
		"file:///etc/passwd",
		"ftp://youtube.com/video",
		"https://tiktok.com/@user/video/123",
	}

	for _, u := range invalid {
		_, err := ValidateYouTubeURL(u)
		if err == nil {
			t.Errorf("ValidateYouTubeURL(%q) expected error, got nil", u)
		}
	}
}

func TestSecurity_ValidateTikTokURL(t *testing.T) {
	valid := []string{
		"https://www.tiktok.com/@user/video/123456789",
		"https://vm.tiktok.com/ZMxxxxxx/",
		"https://vt.tiktok.com/ZSxxxxxx/",
	}

	for _, u := range valid {
		_, err := ValidateTikTokURL(u)
		if err != nil {
			t.Errorf("ValidateTikTokURL(%q) unexpected error: %v", u, err)
		}
	}

	invalid := []string{
		"https://tiktok.com.attacker.com/video/123",
		"http://169.254.169.254/latest/meta-data/",
		"http://192.168.1.50/video",
	}

	for _, u := range invalid {
		_, err := ValidateTikTokURL(u)
		if err == nil {
			t.Errorf("ValidateTikTokURL(%q) expected error, got nil", u)
		}
	}
}

func TestSecurity_ValidateDouyinURL(t *testing.T) {
	valid := []string{
		"https://www.douyin.com/video/7345678901234567890",
		"https://v.douyin.com/ieGsqb4/",
		"https://www.iesdouyin.com/share/video/7345678901234567890/",
		"https://douyin.com/note/7345678901234567890",
	}

	for _, u := range valid {
		_, err := ValidateDouyinURL(u)
		if err != nil {
			t.Errorf("ValidateDouyinURL(%q) unexpected error: %v", u, err)
		}
	}

	invalid := []string{
		"https://douyin.com.attacker.com/video/123",
		"http://169.254.169.254/latest/meta-data/",
		"http://192.168.1.50/video",
		"https://youtube.com/watch?v=123",
	}

	for _, u := range invalid {
		_, err := ValidateDouyinURL(u)
		if err == nil {
			t.Errorf("ValidateDouyinURL(%q) expected error, got nil", u)
		}
	}
}

func TestDouyin_Supports(t *testing.T) {
	dy := NewDouyinDownloader("", "", "")
	if !dy.Supports("https://v.douyin.com/ieGsqb4/") {
		t.Errorf("DouyinDownloader should support v.douyin.com short link")
	}
	if !dy.Supports("https://www.douyin.com/video/7345678901234567890") {
		t.Errorf("DouyinDownloader should support douyin.com video link")
	}
	if dy.Supports("https://www.youtube.com/watch?v=123") {
		t.Errorf("DouyinDownloader should NOT support youtube URL")
	}
}

func TestYouTube_CanonicalURL(t *testing.T) {
	yt := NewYouTubeDownloader("", "", "")
	ctx := context.Background()

	// Short URL youtu.be
	c1, err := yt.CanonicalURL(ctx, "https://youtu.be/dQw4w9WgXcQ")
	if err != nil || c1 != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Errorf("CanonicalURL youtu.be failed: %s, %v", c1, err)
	}

	// Shorts URL
	c2, err := yt.CanonicalURL(ctx, "https://www.youtube.com/shorts/abcdef12345")
	if err != nil || c2 != "https://www.youtube.com/watch?v=abcdef12345" {
		t.Errorf("CanonicalURL shorts failed: %s, %v", c2, err)
	}

	// Playlist URL with video ID: should keep only video ID
	c3, err := yt.CanonicalURL(ctx, "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PL123456")
	if err != nil || c3 != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Errorf("CanonicalURL playlist failed: %s, %v", c3, err)
	}
}

func TestStorage_SanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"../../etc/passwd", "passwd.mp4"},
		{"Hành trình / Việt Nam : 2026?.mp4", "Hanh-trinh-Viet-Nam-2026.mp4"},
		{"Simple Title 123", "Simple-Title-123.mp4"},
	}

	for _, tt := range tests {
		got := storage.SanitizeFilename(tt.input)
		if got == "" || got[0] == '/' || got[0] == '.' {
			t.Errorf("SanitizeFilename(%q) = %q is unsafe", tt.input, got)
		}
	}
}
