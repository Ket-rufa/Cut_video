package downloader

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalidURL          = errors.New("INVALID_URL")
	ErrUnsupportedPlatform = errors.New("UNSUPPORTED_PLATFORM")
	ErrSSRFDetected        = errors.New("SSRF_ATTEMPT_BLOCKED")
	ErrRedirectLimit       = errors.New("REDIRECT_LIMIT_EXCEEDED")
)

var allowedYouTubeHosts = []string{
	"youtube.com",
	"www.youtube.com",
	"m.youtube.com",
	"music.youtube.com",
	"youtu.be",
}

var allowedTikTokHosts = []string{
	"tiktok.com",
	"www.tiktok.com",
	"vm.tiktok.com",
	"vt.tiktok.com",
	"m.tiktok.com",
}

var allowedDouyinHosts = []string{
	"douyin.com",
	"www.douyin.com",
	"v.douyin.com",
	"iesdouyin.com",
	"www.iesdouyin.com",
	"live.douyin.com",
}

// IsAllowedHost checks if the hostname matches the platform allowlist and prevents spoofing.
func IsAllowedHost(host string, allowedHosts []string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	// Remove port if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	for _, allowed := range allowedHosts {
		if host == allowed {
			return true
		}
		// Allow strictly subdomains: e.g. foo.youtube.com, but NEVER youtube.com.fake.com
		if strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}

// IsBlockedIP checks if an IP is loopback, private, link-local, or cloud metadata.
func IsBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	// Loopback (127.0.0.0/8, ::1)
	if ip.IsLoopback() {
		return true
	}

	// Unspecified (0.0.0.0, ::)
	if ip.IsUnspecified() {
		return true
	}

	// Link-local (169.254.0.0/16, fe80::/10)
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Private IP (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7)
	if ip.IsPrivate() {
		return true
	}

	// Explicit metadata service IP check
	if ip.String() == "169.254.169.254" {
		return true
	}

	return false
}

// ValidateTargetURL verifies URL scheme, allowlisted host, and resolves DNS to protect against SSRF.
func ValidateTargetURL(rawURL string, allowedHosts []string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: parse error: %v", ErrInvalidURL, err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("%w: only http/https allowed, got %s", ErrInvalidURL, scheme)
	}

	hostname := strings.ToLower(u.Hostname())
	if hostname == "" {
		return nil, fmt.Errorf("%w: missing hostname", ErrInvalidURL)
	}

	// Check domain allowlist
	if !IsAllowedHost(hostname, allowedHosts) {
		return nil, fmt.Errorf("%w: host %s is not in domain allowlist", ErrUnsupportedPlatform, hostname)
	}

	// Check DNS resolution against SSRF
	ips, err := net.LookupIP(hostname)
	if err != nil {
		// If DNS lookup fails, check if hostname itself is blocked string
		if hostname == "localhost" || strings.HasPrefix(hostname, "127.") || strings.HasPrefix(hostname, "10.") {
			return nil, fmt.Errorf("%w: blocked host: %s", ErrSSRFDetected, hostname)
		}
		// For valid domains in allowlist, network glitch can occur, but if it resolves, check IPs
	} else {
		for _, ip := range ips {
			if IsBlockedIP(ip) {
				return nil, fmt.Errorf("%w: host %s resolved to blocked IP %s", ErrSSRFDetected, hostname, ip.String())
			}
		}
	}

	return u, nil
}

// ValidateYouTubeURL verifies YouTube URL safety.
func ValidateYouTubeURL(rawURL string) (*url.URL, error) {
	return ValidateTargetURL(rawURL, allowedYouTubeHosts)
}

// ValidateTikTokURL verifies TikTok URL safety.
func ValidateTikTokURL(rawURL string) (*url.URL, error) {
	return ValidateTargetURL(rawURL, allowedTikTokHosts)
}

// ValidateDouyinURL verifies Douyin URL safety.
func ValidateDouyinURL(rawURL string) (*url.URL, error) {
	return ValidateTargetURL(rawURL, allowedDouyinHosts)
}

// ResolveRedirectsSafely follows HTTP redirects hop-by-hop with strict SSRF validation at every step.
func ResolveRedirectsSafely(ctx context.Context, startURL string, allowedHosts []string, maxRedirects int) (string, error) {
	if maxRedirects <= 0 {
		maxRedirects = 5
	}

	currentURL := startURL
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Do not follow redirects automatically.
			// We handle each hop manually in the loop below to strictly validate target Location.
			return http.ErrUseLastResponse
		},
	}

	for hop := 0; hop < maxRedirects; hop++ {
		u, err := ValidateTargetURL(currentURL, allowedHosts)
		if err != nil {
			return "", fmt.Errorf("hop %d validation failed: %w", hop, err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodHead, u.String(), nil)
		if err != nil {
			return "", fmt.Errorf("create head request: %w", err)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			// Try GET if HEAD is not supported by endpoint
			req.Method = http.MethodGet
			resp, err = client.Do(req)
			if err != nil {
				return currentURL, nil // Return last valid URL if network ends
			}
		}
		resp.Body.Close()

		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			loc := resp.Header.Get("Location")
			if loc == "" {
				return currentURL, nil
			}

			// Parse Location relative to current URL
			locURL, err := url.Parse(loc)
			if err != nil {
				return "", fmt.Errorf("%w: invalid redirect location: %v", ErrInvalidURL, err)
			}
			resolvedLoc := u.ResolveReference(locURL).String()

			// Validate next hop target BEFORE assigning
			if _, err := ValidateTargetURL(resolvedLoc, allowedHosts); err != nil {
				return "", fmt.Errorf("hop %d redirect target blocked: %w", hop+1, err)
			}

			currentURL = resolvedLoc
			continue
		}

		// Non-redirect response reached canonical endpoint
		return currentURL, nil
	}

	return "", fmt.Errorf("%w: exceeded %d redirects", ErrRedirectLimit, maxRedirects)
}
