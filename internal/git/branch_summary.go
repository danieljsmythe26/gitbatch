package git

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// BranchSummary is the branch data needed by the overview branches panel.
type BranchSummary struct {
	Name         string
	Current      bool
	Dirty        bool
	Upstream     string
	UpstreamGone bool
	Ahead        int
	Behind       int
	Merged       bool
	Worktree     bool
	AgeDays      int
}

// BranchSummaries returns local branches with lightweight status metadata.
func (r *Repository) BranchSummaries(now time.Time) ([]*BranchSummary, error) {
	summaries, err := r.branchSummariesFromRefs(now)
	if err != nil {
		return nil, err
	}

	mergedBranches := r.mergedBranchNames()
	worktreeBranches := r.worktreeBranchNames()
	currentBranch := ""
	if r.State != nil && r.State.Branch != nil {
		currentBranch = r.State.Branch.Name
	}

	for _, summary := range summaries {
		summary.Current = summary.Name == currentBranch
		summary.Dirty = summary.Current && r.State != nil && r.State.Branch != nil && !r.State.Branch.Clean
		summary.Merged = !summary.Current && mergedBranches[summary.Name]
		summary.Worktree = !summary.Current && worktreeBranches[summary.Name]
	}

	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].Current != summaries[j].Current {
			return summaries[i].Current
		}
		return summaries[i].Name < summaries[j].Name
	})

	return summaries, nil
}

func (r *Repository) branchSummariesFromRefs(now time.Time) ([]*BranchSummary, error) {
	args := []string{
		"for-each-ref",
		"--format=%(refname:short)%00%(upstream:short)%00%(upstream:track)%00%(committerdate:unix)",
		"refs/heads",
	}
	out, err := r.gitOutput(args...)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	summaries := make([]*BranchSummary, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\x00")
		for len(parts) < 4 {
			parts = append(parts, "")
		}
		ahead, behind, gone := parseBranchTrack(parts[2])
		ageDays := branchAgeDays(now, parts[3])
		summaries = append(summaries, &BranchSummary{
			Name:         parts[0],
			Upstream:     parts[1],
			UpstreamGone: gone,
			Ahead:        ahead,
			Behind:       behind,
			AgeDays:      ageDays,
		})
	}
	return summaries, nil
}

func parseBranchTrack(track string) (int, int, bool) {
	track = strings.Trim(track, "[] ")
	if track == "" {
		return 0, 0, false
	}
	if track == "gone" {
		return 0, 0, true
	}

	var ahead, behind int
	for _, part := range strings.Split(track, ",") {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) != 2 {
			continue
		}
		count, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		switch fields[0] {
		case "ahead":
			ahead = count
		case "behind":
			behind = count
		}
	}
	return ahead, behind, false
}

func branchAgeDays(now time.Time, unixSeconds string) int {
	seconds, err := strconv.ParseInt(strings.TrimSpace(unixSeconds), 10, 64)
	if err != nil || seconds <= 0 {
		return 0
	}
	age := now.Sub(time.Unix(seconds, 0))
	if age < 0 {
		return 0
	}
	return int(age.Hours() / 24)
}

func (r *Repository) mergedBranchNames() map[string]bool {
	branches := make(map[string]bool)
	out, err := r.gitOutput("branch", "--merged")
	if err != nil {
		return branches
	}
	for _, line := range strings.Split(out, "\n") {
		name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "*"))
		if name != "" {
			branches[name] = true
		}
	}
	return branches
}

func (r *Repository) worktreeBranchNames() map[string]bool {
	branches := make(map[string]bool)
	out, err := r.gitOutput("worktree", "list", "--porcelain")
	if err != nil {
		return branches
	}
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "branch refs/heads/") {
			continue
		}
		name := strings.TrimPrefix(line, "branch refs/heads/")
		if name != "" {
			branches[name] = true
		}
	}
	return branches
}

func (r *Repository) gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.AbsPath
	out, err := cmd.CombinedOutput()
	return string(out), err
}
