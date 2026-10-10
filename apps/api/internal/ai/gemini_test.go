package ai

import (
	"testing"

	"social-video-cutter/api/internal/domain"
)

func TestParseSRTWithoutBlankLines(t *testing.T) {
	srt := `1
00:00:00,320 --> 00:00:03,420
Người cha chưa bao giờ mơ thấy điều này.
2
00:00:03,660 --> 00:00:05,980
Đó là máy quét nhận dạng khuôn mặt.
3
00:00:09,000 --> 00:00:11,100
Biệt thự nằm ngay trước cửa.`

	cues := ParseSRT(srt, 90000)
	if len(cues) != 3 {
		t.Fatalf("expected 3 cues, got %d: %#v", len(cues), cues)
	}
	if cues[0].StartMs != 90320 || cues[0].EndMs != 93420 {
		t.Fatalf("unexpected first cue timing: %#v", cues[0])
	}
	if cues[0].Text != "Người cha chưa bao giờ mơ thấy điều này." {
		t.Fatalf("unexpected first cue text: %q", cues[0].Text)
	}
	if cues[1].Text != "Đó là máy quét nhận dạng khuôn mặt." {
		t.Fatalf("unexpected second cue text: %q", cues[1].Text)
	}
}

func TestParseSRTWithBlankLinesAndMarkdownFence(t *testing.T) {
	srt := "```srt\n1\n00:00:00.000 --> 00:00:02.500\nDòng thứ nhất.\n\n2\n00:00:02,700 --> 00:00:05,000\nDòng thứ hai.\n```"

	cues := ParseSRT(srt, 0)
	if len(cues) != 2 {
		t.Fatalf("expected 2 cues, got %d: %#v", len(cues), cues)
	}
	if cues[0].Text != "Dòng thứ nhất." || cues[1].Text != "Dòng thứ hai." {
		t.Fatalf("unexpected cue texts: %#v", cues)
	}
}

func TestValidateSRTChunkRejectsOversizedCue(t *testing.T) {
	srt := "1\n00:00:00,000 --> 00:01:35,000\nMột câu bị gom quá dài."
	if err := validateSRTChunk(srt, 100000); err == nil {
		t.Fatal("expected an oversized cue validation error")
	}
}

func TestValidateSRTChunkAcceptsShortSentenceCues(t *testing.T) {
	srt := "1\n00:00:00,000 --> 00:00:02,500\nCâu thứ nhất.\n2\n00:00:02,700 --> 00:00:05,000\nCâu thứ hai."
	if err := validateSRTChunk(srt, 6000); err != nil {
		t.Fatalf("expected valid cues, got %v", err)
	}
}

func TestFindBoundaryGaps(t *testing.T) {
	cues := []domain.SubtitleCue{
		{StartMs: 900000, EndMs: 921000, Text: "Trước ranh giới."},
		{StartMs: 960000, EndMs: 963000, Text: "Sau ranh giới."},
		{StartMs: 1180000, EndMs: 1190000, Text: "Không có khoảng trống lớn."},
		{StartMs: 1200000, EndMs: 1203000, Text: "Đoạn tiếp theo."},
	}

	gaps := findBoundaryGaps(cues, []int64{960000, 1200000}, 15000, 60000)
	if len(gaps) != 1 || gaps[0].start != 921000 || gaps[0].end != 960000 {
		t.Fatalf("unexpected boundary gaps: %#v", gaps)
	}
}

func TestSplitLongCueWithoutPunctuation(t *testing.T) {
	cue := domain.SubtitleCue{StartMs: 0, EndMs: 18000, Text: "một câu thoại khá dài cần được chia thành nhiều phần ngắn"}
	parts := SplitLongCueIfNeeded(cue)
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d: %#v", len(parts), parts)
	}
	for _, part := range parts {
		if part.EndMs-part.StartMs > 7000 {
			t.Fatalf("part is too long: %#v", part)
		}
	}
}
