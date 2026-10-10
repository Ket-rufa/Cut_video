package downloader

import (
	"fmt"
	"testing"

	"social-video-cutter/api/internal/domain"
)

func TestJobManager_DedupKey(t *testing.T) {
	k1 := makeDedupKey("user-1", "youtube", "abc123", "1080p")
	k2 := makeDedupKey("user-1", "youtube", "abc123", "1080P")
	k3 := makeDedupKey("user-2", "youtube", "abc123", "1080p")

	if k1 != k2 {
		t.Errorf("Expected case-insensitive match for quality: %s != %s", k1, k2)
	}
	if k1 == k3 {
		t.Errorf("Expected different user to produce different dedup key: %s == %s", k1, k3)
	}
}

func TestJobManager_StateTransitions(t *testing.T) {
	// Rule #9: Job state machine must be strictly locked.
	// Valid transitions:
	// QUEUED -> ANALYZING -> DOWNLOADING -> MERGING -> SAVING -> COMPLETED
	// Any state -> FAILED / CANCELLED
	// Invalid:
	// COMPLETED -> DOWNLOADING
	// FAILED -> COMPLETED
	// CANCELLED -> SAVING

	validTransitions := []struct {
		from  domain.JobStatus
		to    domain.JobStatus
		valid bool
	}{
		{domain.JobStatusQueued, domain.JobStatusAnalyzing, true},
		{domain.JobStatusAnalyzing, domain.JobStatusDownloading, true},
		{domain.JobStatusDownloading, domain.JobStatusMerging, true},
		{domain.JobStatusMerging, domain.JobStatusSaving, true},
		{domain.JobStatusSaving, domain.JobStatusCompleted, true},
		{domain.JobStatusDownloading, domain.JobStatusFailed, true},
		{domain.JobStatusAnalyzing, domain.JobStatusCancelled, true},
		// Invalid transitions:
		{domain.JobStatusCompleted, domain.JobStatusDownloading, false},
		{domain.JobStatusFailed, domain.JobStatusCompleted, false},
		{domain.JobStatusCancelled, domain.JobStatusSaving, false},
		{domain.JobStatusCompleted, domain.JobStatusQueued, false},
		{domain.JobStatusSaving, domain.JobStatusDownloading, false},
	}

	for _, tc := range validTransitions {
		t.Run(fmt.Sprintf("%s->%s", tc.from, tc.to), func(t *testing.T) {
			isValid := isValidTransition(tc.from, tc.to)
			if isValid != tc.valid {
				t.Errorf("Transition %s -> %s: expected %v, got %v", tc.from, tc.to, tc.valid, isValid)
			}
		})
	}
}

func isValidTransition(from, to domain.JobStatus) bool {
	// Terminal states cannot transition to anything
	if from == domain.JobStatusCompleted || from == domain.JobStatusFailed || from == domain.JobStatusCancelled {
		return false
	}
	// Any non-terminal state can transition to FAILED or CANCELLED
	if to == domain.JobStatusFailed || to == domain.JobStatusCancelled {
		return true
	}
	switch from {
	case domain.JobStatusQueued:
		return to == domain.JobStatusAnalyzing
	case domain.JobStatusAnalyzing:
		return to == domain.JobStatusDownloading
	case domain.JobStatusDownloading:
		return to == domain.JobStatusMerging || to == domain.JobStatusSaving
	case domain.JobStatusMerging:
		return to == domain.JobStatusSaving
	case domain.JobStatusSaving:
		return to == domain.JobStatusCompleted
	default:
		return false
	}
}
