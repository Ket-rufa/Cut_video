package downloader

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSignedDouyinDetailURL(t *testing.T) {
	got := signedDouyinDetailURL("7588437942075510035", "abc", time.Unix(1700000000, 0))
	if !strings.Contains(got, "timestamp=1700000000") {
		t.Fatalf("signed URL is missing timestamp: %s", got)
	}
	if !strings.HasSuffix(got, "x-secsdk-web-signature=d7694d17ae11ffc0703a71f67c904b14") {
		t.Fatalf("unexpected SecSDK signature: %s", got)
	}
}

func TestReadDouyinCookiesSkipsMalformedAndExpiredRows(t *testing.T) {
	path := t.TempDir() + "/cookies.txt"
	content := "# Netscape HTTP Cookie File\n" +
		".douyin.com\tTRUE\t/\tTRUE\t4102444800\tUIFID\tfingerprint\n" +
		"www.douyin.com\tFALSE\t/\tFALSE\t4102444800\tsessionid\tvalid-session\n" +
		"www.douyin.com\tFALSE\t/\tFALSE\t1\texpired\told\n" +
		"www.douyin.com\tFALSE\t/\tFALSE\t0\t\tdouyin.com\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	header, uifid, err := readDouyinCookies(path)
	if err != nil {
		t.Fatal(err)
	}
	if uifid != "fingerprint" || header != "UIFID=fingerprint; sessionid=valid-session" {
		t.Fatalf("unexpected parsed cookies: uifid=%q header=%q", uifid, header)
	}
}

func TestDouyin_EnsureCookies(t *testing.T) {
	d := NewDouyinDownloader("", "", "")
	tmpCookie := t.TempDir() + "/cookies.txt"
	err := d.ensureDouyinCookieFile(tmpCookie)
	if err != nil {
		t.Fatalf("ensureDouyinCookieFile failed: %v", err)
	}
}

func TestDouyin_Analyze(t *testing.T) {
	d := NewDouyinDownloader("yt-dlp", "ffmpeg", "ffprobe")
	ctx := context.Background()

	url := "https://www.douyin.com/jingxuan?modal_id=7689326480735620402"
	res, err := d.Analyze(ctx, url)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if res.VideoID != "7689326480735620402" {
		t.Errorf("VideoID = %q, want %q", res.VideoID, "7689326480735620402")
	}
	if len(res.Formats) == 0 {
		t.Errorf("expected formats, got empty")
	}
}

func TestDouyin_RealDownload(t *testing.T) {
	d := NewDouyinDownloader("yt-dlp", "ffmpeg", "ffprobe")
	ctx := context.Background()

	tmpDir := t.TempDir()
	opt := DownloadOptions{
		SourceURL: "https://www.douyin.com/video/7588437942075510035",
		Quality:   "original",
		TempDir:   tmpDir,
	}

	res, err := d.Download(ctx, opt, func(p DownloadProgress) {
		t.Logf("Douyin progress: %s - %d%%", p.Status, p.Progress)
	})
	if err != nil {
		t.Fatalf("Douyin Download failed: %v", err)
	}

	t.Logf("Douyin Download Success: File=%s, Size=%d, Res=%dx%d", res.FileName, res.FileSize, res.Width, res.Height)
}
