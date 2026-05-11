package git

import (
	"strconv"
	"testing"
	"time"
)

func TestParseBranchTrack(t *testing.T) {
	tests := []struct {
		name       string
		track      string
		wantAhead  int
		wantBehind int
		wantGone   bool
	}{
		{name: "empty", track: "", wantAhead: 0, wantBehind: 0, wantGone: false},
		{name: "ahead", track: "[ahead 3]", wantAhead: 3, wantBehind: 0, wantGone: false},
		{name: "behind", track: "[behind 5]", wantAhead: 0, wantBehind: 5, wantGone: false},
		{name: "diverged", track: "[ahead 3, behind 5]", wantAhead: 3, wantBehind: 5, wantGone: false},
		{name: "gone", track: "[gone]", wantAhead: 0, wantBehind: 0, wantGone: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ahead, behind, gone := parseBranchTrack(tt.track)
			if ahead != tt.wantAhead || behind != tt.wantBehind || gone != tt.wantGone {
				t.Fatalf("parseBranchTrack(%q) = (%d, %d, %t), want (%d, %d, %t)",
					tt.track, ahead, behind, gone, tt.wantAhead, tt.wantBehind, tt.wantGone)
			}
		})
	}
}

func TestBranchAgeDays(t *testing.T) {
	now := time.Unix(200000, 0)
	commitTime := now.Add(-49 * time.Hour).Unix()

	if got := branchAgeDays(now, "not-a-time"); got != 0 {
		t.Fatalf("invalid time age = %d, want 0", got)
	}
	if got := branchAgeDays(now, "200001"); got != 0 {
		t.Fatalf("future age = %d, want 0", got)
	}
	if got := branchAgeDays(now, "235200"); got != 0 {
		t.Fatalf("future timestamp age = %d, want 0", got)
	}
	if got := branchAgeDays(now, strconv.FormatInt(commitTime, 10)); got != 2 {
		t.Fatalf("age = %d, want 2", got)
	}
}
