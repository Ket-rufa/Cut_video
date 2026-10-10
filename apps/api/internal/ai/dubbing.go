package ai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"social-video-cutter/api/internal/config"
	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/repository"
)

type DubbingService struct {
	cfg        *config.Config
	subRepo    repository.SubtitleRepository
	httpClient *http.Client
}

func NewDubbingService(cfg *config.Config, subRepo repository.SubtitleRepository) *DubbingService {
	return &DubbingService{
		cfg:        cfg,
		subRepo:    subRepo,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SynthesizeRawTTS downloads Vietnamese speech MP3 bytes from Google TTS.
func (s *DubbingService) SynthesizeRawTTS(ctx context.Context, text string) ([]byte, error) {
	text = SanitizeTTSText(text)
	if text == "" {
		return nil, fmt.Errorf("văn bản rỗng")
	}

	// Split if longer than 180 characters to comply with Google TTS URL length limits
	if len([]rune(text)) > 150 {
		parts := splitSentence(text, 140)
		var combined bytes.Buffer
		for _, part := range parts {
			b, err := s.fetchTTSChunk(ctx, part)
			if err != nil {
				return nil, err
			}
			combined.Write(b)
		}
		return combined.Bytes(), nil
	}

	return s.fetchTTSChunk(ctx, text)
}

var (
	ttsNumberedTimestampRegex = regexp.MustCompile(`\s+\d+[.)]?\s+\d{2}:\d{2}:\d{2}[,.]\d{3}\s*-->\s*\d{2}:\d{2}:\d{2}[,.]\d{3}\s*`)
	ttsTimestampRegex         = regexp.MustCompile(`\d{2}:\d{2}:\d{2}[,.]\d{3}\s*-->\s*\d{2}:\d{2}:\d{2}[,.]\d{3}`)
	ttsStandaloneIndexRegex   = regexp.MustCompile(`^\d+[.)]?$`)
)

// SanitizeTTSText prevents malformed SRT content from being spoken aloud.
func SanitizeTTSText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = ttsNumberedTimestampRegex.ReplaceAllString(text, " ")
	text = ttsTimestampRegex.ReplaceAllString(text, " ")

	var spokenLines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || ttsStandaloneIndexRegex.MatchString(line) {
			continue
		}
		spokenLines = append(spokenLines, line)
	}
	return strings.Join(strings.Fields(strings.Join(spokenLines, " ")), " ")
}

func (s *DubbingService) fetchTTSChunk(ctx context.Context, text string) ([]byte, error) {
	ttsURL := fmt.Sprintf("https://translate.google.com/translate_tts?ie=UTF-8&q=%s&tl=vi&client=tw-ob", url.QueryEscape(text))
	req, err := http.NewRequestWithContext(ctx, "GET", ttsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lỗi kết nối TTS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TTS trả về HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func splitSentence(text string, maxLen int) []string {
	var chunks []string
	words := strings.Fields(text)
	var current strings.Builder

	for _, w := range words {
		if current.Len()+len(w)+1 > maxLen && current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}
	return chunks
}

func (s *DubbingService) probeDuration(ctx context.Context, audioPath string) float64 {
	cmd := exec.CommandContext(ctx, s.cfg.Downloader.FFprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	dur, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return dur
}

// VoiceOption represents a selectable TTS voice profile.
type VoiceOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	Description string `json:"description"`
	Tag         string `json:"tag"`
	Icon        string `json:"icon"`
	SampleText  string `json:"sample_text"`
}

// AvailableVoices lists all supported TTS voice profiles for Vietnamese dubbing.
var AvailableVoices = []VoiceOption{
	{
		ID:          "male_deep",
		Name:        "Nam Trầm Ấm (Review Phim)",
		Gender:      "male",
		Description: "Sâu lắng, ấm áp, kịch tính, chuẩn phong cách review phim ngắn TikTok",
		Tag:         "Khuyên dùng",
		Icon:        "👨",
		SampleText:  "Hôm nay, tôi sẽ tóm tắt cho bạn một bộ phim vô cùng lôi cuốn và kịch tính.",
	},
	{
		ID:          "female_standard",
		Name:        "Nữ Tiêu Chuẩn (Google)",
		Gender:      "female",
		Description: "Tự nhiên, trong trẻo, phổ biến và dễ nghe",
		Tag:         "Mặc định",
		Icon:        "👩",
		SampleText:  "Xin chào, đây là giọng đọc nữ tiêu chuẩn dành cho video của bạn.",
	},
	{
		ID:          "female_young",
		Name:        "Nữ Trẻ Trung (Vlog / Ngọt Ngào)",
		Gender:      "female",
		Description: "Ngọt ngào, nhẹ nhàng, thích hợp vlog đời sống & ẩm thực",
		Tag:         "Vlog",
		Icon:        "👧",
		SampleText:  "Chào mọi người, cùng mình trải nghiệm một ngày thật đáng nhớ nhé!",
	},
	{
		ID:          "male_news",
		Name:        "Nam Năng Động (Tin Tức / Phóng Sự)",
		Gender:      "male",
		Description: "Rõ ràng, dứt khoát, phong cách thời sự và tài liệu",
		Tag:         "Tin tức",
		Icon:        "🎙️",
		SampleText:  "Bản tin cập nhật những diễn biến mới nhất ngay trong hôm nay.",
	},
	{
		ID:          "male_story",
		Name:        "Nam Kể Chuyện (Đêm Khuya / Trầm)",
		Gender:      "male",
		Description: "Truyền cảm, vang ấm, thích hợp kể chuyện & podcast",
		Tag:         "Podcast",
		Icon:        "📻",
		SampleText:  "Trong một đêm mưa lạnh, câu chuyện bí ẩn bắt đầu hé lộ.",
	},
	{
		ID:          "robot",
		Name:        "Robot AI (Công Nghệ Sci-Fi)",
		Gender:      "other",
		Description: "Âm thanh điện tử, phong cách khoa học viễn tưởng",
		Tag:         "AI Tech",
		Icon:        "🤖",
		SampleText:  "Hệ thống trí tuệ nhân tạo đã khởi tạo hoàn tất và sẵn sàng hoạt động.",
	},
	{
		ID:          "fun_chipmunk",
		Name:        "Hoạt Hình / Nhí Nhảnh",
		Gender:      "other",
		Description: "Vui tươi, hài hước, tạo hiệu ứng giải trí thú vị",
		Tag:         "Hài hước",
		Icon:        "🐿️",
		SampleText:  "Ôi vui quá đi mất, cùng quẩy lên nào các bạn ơi!",
	},
}

// VoiceFilters maps each voice ID to its dedicated FFmpeg audio filter chain.
var VoiceFilters = map[string]string{
	"female_standard": "highpass=f=80,equalizer=f=3000:width_type=o:width=1:g=1",
	"male_deep":       "asetrate=24000*0.80,aresample=24000,atempo=1/0.80,equalizer=f=120:width_type=o:width=1.5:g=6,equalizer=f=2800:width_type=o:width=1:g=-2",
	"female_young":    "asetrate=24000*1.15,aresample=24000,atempo=1/1.15,equalizer=f=3200:width_type=o:width=1.5:g=3",
	"male_news":       "asetrate=24000*0.90,aresample=24000,atempo=1/0.90,equalizer=f=200:width_type=o:width=1:g=3,equalizer=f=2500:width_type=o:width=1:g=2",
	"male_story":      "asetrate=24000*0.86,aresample=24000,atempo=1/0.86,equalizer=f=100:width_type=o:width=2:g=4,aecho=0.8:0.88:30:0.2",
	"robot":           "flanger=speed=0.5:depth=2,tremolo=f=40:d=0.35",
	"fun_chipmunk":    "asetrate=24000*1.30,aresample=24000,atempo=1/1.30",
}

// SynthesizeSampleVoice creates a short preview audio for the chosen voice.
func (s *DubbingService) SynthesizeSampleVoice(ctx context.Context, voice string) ([]byte, error) {
	sampleText := "Xin chào các bạn, đây là giọng đọc AI hỗ trợ lồng tiếng video."
	for _, v := range AvailableVoices {
		if v.ID == voice && v.SampleText != "" {
			sampleText = v.SampleText
			break
		}
	}

	rawBytes, err := s.SynthesizeRawTTS(ctx, sampleText)
	if err != nil {
		return nil, err
	}

	filter, hasVoice := VoiceFilters[voice]
	if !hasVoice || filter == "" {
		return rawBytes, nil
	}

	tempDir := filepath.Join(s.cfg.Storage.Root, "temp")
	_ = os.MkdirAll(tempDir, 0755)
	tempIn := filepath.Join(tempDir, fmt.Sprintf("sample_in_%s_%d.mp3", voice, time.Now().UnixNano()))
	tempOut := filepath.Join(tempDir, fmt.Sprintf("sample_out_%s_%d.mp3", voice, time.Now().UnixNano()))
	defer os.Remove(tempIn)
	defer os.Remove(tempOut)

	if err := os.WriteFile(tempIn, rawBytes, 0644); err != nil {
		return rawBytes, nil
	}

	cmd := exec.CommandContext(ctx, s.cfg.Downloader.FFmpegPath,
		"-y",
		"-i", tempIn,
		"-af", filter,
		"-ar", "24000",
		"-b:a", "64k",
		tempOut,
	)
	if err := cmd.Run(); err != nil {
		return rawBytes, nil
	}

	voicedBytes, err := os.ReadFile(tempOut)
	if err != nil {
		return rawBytes, nil
	}
	return voicedBytes, nil
}

// SynthesizeSyncedCueTTS generates TTS audio, applies voice character styling, and stretches/compresses tempo with FFmpeg.
func (s *DubbingService) SynthesizeSyncedCueTTS(ctx context.Context, text string, targetDurSec float64, outputPath string, voice string) error {
	rawBytes, err := s.SynthesizeRawTTS(ctx, text)
	if err != nil {
		return err
	}

	tempRaw := outputPath + ".raw.mp3"
	if err := os.WriteFile(tempRaw, rawBytes, 0644); err != nil {
		return err
	}
	defer os.Remove(tempRaw)

	// Step 1: Apply Voice Profile Filter (if specified and recognized)
	voiceSource := tempRaw
	if voice == "" {
		voice = "male_deep"
	}
	filter, hasVoice := VoiceFilters[voice]
	tempVoiced := outputPath + ".voice.mp3"
	if hasVoice && filter != "" {
		cmdVoice := exec.CommandContext(ctx, s.cfg.Downloader.FFmpegPath,
			"-y",
			"-i", tempRaw,
			"-af", filter,
			"-ar", "24000",
			"-b:a", "64k",
			tempVoiced,
		)
		if err := cmdVoice.Run(); err == nil {
			voiceSource = tempVoiced
			defer os.Remove(tempVoiced)
		}
	}

	actualDur := s.probeDuration(ctx, voiceSource)
	if actualDur <= 0 || targetDurSec <= 0 {
		b, rErr := os.ReadFile(voiceSource)
		if rErr == nil {
			return os.WriteFile(outputPath, b, 0644)
		}
		return os.WriteFile(outputPath, rawBytes, 0644)
	}

	// Step 2: Calculate tempo multiplier to fit target rhythm
	tempo := actualDur / targetDurSec
	// Clamp tempo between 0.65x and 1.8x to preserve speech clarity
	if tempo < 0.65 {
		tempo = 0.65
	}
	if tempo > 1.8 {
		tempo = 1.8
	}

	// Build atempo filter chain (atempo only accepts 0.5 to 2.0 per filter)
	var atempoFilter string
	if tempo > 2.0 {
		atempoFilter = fmt.Sprintf("atempo=2.0,atempo=%.3f", tempo/2.0)
	} else if tempo < 0.5 {
		atempoFilter = fmt.Sprintf("atempo=0.5,atempo=%.3f", tempo/0.5)
	} else {
		atempoFilter = fmt.Sprintf("atempo=%.3f", tempo)
	}

	cmd := exec.CommandContext(ctx, s.cfg.Downloader.FFmpegPath,
		"-y",
		"-i", voiceSource,
		"-filter:a", atempoFilter,
		"-ar", "24000",
		"-b:a", "64k",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		b, _ := os.ReadFile(voiceSource)
		_ = os.WriteFile(outputPath, b, 0644)
		return fmt.Errorf("ffmpeg tempo adjust failed: %v (output: %s)", err, string(out))
	}

	return nil
}

// GenerateDubbingTrack generates a synchronized audio track covering the video/clip range using the selected voice.
func (s *DubbingService) GenerateDubbingTrack(ctx context.Context, video *domain.Video, cues []domain.SubtitleCue, startMs, endMs int64, bgVolume float64, voice string) (*domain.VideoDubbing, error) {
	if len(cues) == 0 {
		return nil, fmt.Errorf("không có phụ đề nào để lồng tiếng")
	}

	if endMs <= startMs {
		if video.DurationMs > 0 {
			endMs = video.DurationMs
		} else {
			endMs = cues[len(cues)-1].EndMs
		}
	}

	// Filter cues inside [startMs, endMs]
	var inRange []domain.SubtitleCue
	for _, c := range cues {
		if c.EndMs >= startMs && c.StartMs <= endMs {
			inRange = append(inRange, c)
		}
	}
	if len(inRange) == 0 {
		inRange = cues
	}

	dubbingDir := filepath.Join(s.cfg.Storage.Root, "dubbing")
	_ = os.MkdirAll(dubbingDir, 0755)

	tempDir := filepath.Join(s.cfg.Storage.Root, "temp", fmt.Sprintf("dub_%s_%d", video.ID, time.Now().UnixNano()))
	_ = os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	type cueResult struct {
		index    int
		cue      domain.SubtitleCue
		filePath string
		err      error
	}

	// Synthesize cues in parallel (concurrency 4)
	cueResults := make([]cueResult, len(inRange))
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup

	for i, c := range inRange {
		wg.Add(1)
		go func(idx int, cue domain.SubtitleCue) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			targetSec := float64(cue.EndMs-cue.StartMs) / 1000.0
			outCuePath := filepath.Join(tempDir, fmt.Sprintf("cue_%d.mp3", cue.Index))
			err := s.SynthesizeSyncedCueTTS(ctx, cue.Text, targetSec, outCuePath, voice)
			cueResults[idx] = cueResult{
				index:    idx,
				cue:      cue,
				filePath: outCuePath,
				err:      err,
			}
		}(i, c)
	}
	wg.Wait()

	// Assemble segments list: silence gaps + cue speech
	var concatItems []string
	curTimeMs := startMs

	for _, cr := range cueResults {
		if cr.err != nil {
			continue
		}

		// Gap before this cue
		gapMs := cr.cue.StartMs - curTimeMs
		if gapMs > 50 {
			gapSec := float64(gapMs) / 1000.0
			silencePath := filepath.Join(tempDir, fmt.Sprintf("silence_%d.mp3", curTimeMs))
			cmdSil := exec.CommandContext(ctx, s.cfg.Downloader.FFmpegPath,
				"-y",
				"-f", "lavfi",
				"-i", "anullsrc=r=24000:cl=mono",
				"-t", fmt.Sprintf("%.3f", gapSec),
				"-b:a", "64k",
				silencePath,
			)
			if err := cmdSil.Run(); err == nil {
				concatItems = append(concatItems, silencePath)
			}
		}

		concatItems = append(concatItems, cr.filePath)
		actualDur := s.probeDuration(ctx, cr.filePath)
		if actualDur > 0 {
			curTimeMs = cr.cue.StartMs + int64(actualDur*1000)
		} else {
			curTimeMs = cr.cue.EndMs
		}
	}

	// Remaining tail silence to match endMs
	if endMs > curTimeMs {
		tailSec := float64(endMs-curTimeMs) / 1000.0
		tailPath := filepath.Join(tempDir, "silence_tail.mp3")
		cmdTail := exec.CommandContext(ctx, s.cfg.Downloader.FFmpegPath,
			"-y",
			"-f", "lavfi",
			"-i", "anullsrc=r=24000:cl=mono",
			"-t", fmt.Sprintf("%.3f", tailSec),
			"-b:a", "64k",
			tailPath,
		)
		if err := cmdTail.Run(); err == nil {
			concatItems = append(concatItems, tailPath)
		}
	}

	if len(concatItems) == 0 {
		return nil, fmt.Errorf("không tạo được phân đoạn âm thanh lồng tiếng")
	}

	// Write concat list file using base filenames (relative to concat.txt itself)
	concatListFile := filepath.Join(tempDir, "concat.txt")
	var concatText strings.Builder
	for _, item := range concatItems {
		concatText.WriteString(fmt.Sprintf("file '%s'\n", filepath.Base(item)))
	}
	if err := os.WriteFile(concatListFile, []byte(concatText.String()), 0644); err != nil {
		return nil, err
	}

	finalAudioFileName := fmt.Sprintf("dubbed_audio_%s.mp3", video.ID)
	finalAudioPath, _ := filepath.Abs(filepath.Join(dubbingDir, finalAudioFileName))

	cmdConcat := exec.CommandContext(ctx, s.cfg.Downloader.FFmpegPath,
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", "concat.txt",
		"-c:a", "libmp3lame",
		"-ar", "24000",
		"-b:a", "64k",
		finalAudioPath,
	)
	cmdConcat.Dir = tempDir
	if out, err := cmdConcat.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("lỗi ghép nối audio lồng tiếng: %v (output: %s)", err, string(out))
	}

	// If source video exists, also mix video with dubbed audio
	finalVideoFileName := fmt.Sprintf("dubbed_video_%s.mp4", video.ID)
	finalVideoPath, _ := filepath.Abs(filepath.Join(dubbingDir, finalVideoFileName))
	hasDubbedVideo := false

	if video.SourcePath != "" {
		srcVideo := video.SourcePath
		if _, err := os.Stat(srcVideo); os.IsNotExist(err) {
			cand := filepath.Join(s.cfg.Storage.Root, srcVideo)
			if _, err := os.Stat(cand); err == nil {
				srcVideo = cand
			} else {
				cleanRoot := filepath.Clean(s.cfg.Storage.Root)
				if strings.HasPrefix(srcVideo, cleanRoot+string(filepath.Separator)) {
					rel := strings.TrimPrefix(srcVideo, cleanRoot+string(filepath.Separator))
					cand2 := filepath.Join(s.cfg.Storage.Root, rel)
					if _, err := os.Stat(cand2); err == nil {
						srcVideo = cand2
					}
				}
			}
		}

		if stat, err := os.Stat(srcVideo); err == nil && !stat.IsDir() {
			if bgVolume <= 0 {
				bgVolume = 0.15
			}
			filterComplex := fmt.Sprintf("[0:a]volume=%.2f[bg];[1:a]volume=1.0[voice];[bg][voice]amix=inputs=2:duration=first[aout]", bgVolume)
			ffmpegMixArgs := []string{"-y"}
			if startMs > 0 {
				ffmpegMixArgs = append(ffmpegMixArgs, "-ss", fmt.Sprintf("%.3f", float64(startMs)/1000.0))
			}
			ffmpegMixArgs = append(ffmpegMixArgs, "-i", srcVideo)
			if endMs > startMs {
				ffmpegMixArgs = append(ffmpegMixArgs, "-t", fmt.Sprintf("%.3f", float64(endMs-startMs)/1000.0))
			}
			ffmpegMixArgs = append(ffmpegMixArgs,
				"-i", finalAudioPath,
				"-filter_complex", filterComplex,
				"-map", "0:v",
				"-map", "[aout]",
				"-c:v", "copy",
				"-c:a", "aac",
				"-b:a", "128k",
				finalVideoPath,
			)

			cmdMix := exec.CommandContext(ctx, s.cfg.Downloader.FFmpegPath, ffmpegMixArgs...)
			if out, err := cmdMix.CombinedOutput(); err == nil {
				hasDubbedVideo = true
			} else {
				fmt.Printf("[MIX DUBBING VIDEO WARNING] %v (output: %s)\n", err, string(out))
			}
		}
	}

	audioURL := fmt.Sprintf("/api/v1/videos/%s/dubbing/audio", video.ID)
	var videoURL string
	if hasDubbedVideo {
		videoURL = fmt.Sprintf("/api/v1/videos/%s/dubbing/video", video.ID)
	}

	result := &domain.VideoDubbing{
		VideoID:    video.ID,
		AudioURL:   audioURL,
		VideoURL:   videoURL,
		CueCount:   len(inRange),
		StartMs:    startMs,
		EndMs:      endMs,
		DurationMs: endMs - startMs,
		Status:     "ready",
		CreatedAt:  time.Now(),
	}

	return result, nil
}
