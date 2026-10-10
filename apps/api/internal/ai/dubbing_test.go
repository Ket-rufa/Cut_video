package ai

import "testing"

func TestSanitizeTTSTextRemovesSRTMetadata(t *testing.T) {
	input := "1\n00:00:00,000 --> 00:00:03,800\nNgười bố nằm mơ cũng không ngờ.\n2\n00:00:03,800 --> 00:00:08,200\nCổng biệt thự rất xa hoa."
	want := "Người bố nằm mơ cũng không ngờ. Cổng biệt thự rất xa hoa."
	if got := SanitizeTTSText(input); got != want {
		t.Fatalf("unexpected sanitized TTS text: %q", got)
	}
}

func TestSanitizeTTSTextRemovesInlineSRTMetadata(t *testing.T) {
	input := "Người bố rất bất ngờ. 2 00:00:03,800 --> 00:00:08,200 Cổng biệt thự rất xa hoa."
	want := "Người bố rất bất ngờ. Cổng biệt thự rất xa hoa."
	if got := SanitizeTTSText(input); got != want {
		t.Fatalf("unexpected sanitized inline TTS text: %q", got)
	}
}
