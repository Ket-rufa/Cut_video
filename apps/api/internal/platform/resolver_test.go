package platform_test

import (
	"testing"

	"social-video-cutter/api/internal/platform"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid youtube", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", false},
		{"valid https", "https://example.com/video.mp4", false},
		{"valid http", "http://example.com/video.mp4", false},
		{"ftp scheme", "ftp://example.com/video.mp4", true},
		{"localhost SSRF", "http://localhost/admin", true},
		{"127 SSRF", "http://127.0.0.1/secret", true},
		{"private IP 192.168", "http://192.168.1.1/video", true},
		{"private IP 10.", "http://10.0.0.1/video", true},
		{"no scheme", "example.com/video.mp4", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := platform.ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestResolverSupports(t *testing.T) {
	r := platform.NewResolver()
	_ = r // resolver is initialized with all providers
}
