package gui

import (
	"strings"
	"testing"

	"github.com/isacikgoz/gitbatch/internal/git"
)

func TestRepositoryTableLineKeepsSeparatorPositionsAligned(t *testing.T) {
	rule := &RepositoryDecorationRules{
		MaxName:      18,
		MaxPushables: 1,
		MaxPullables: 1,
		MaxBranch:    12,
	}

	repo := &git.Repository{
		Name: "agent-config",
		State: &git.RepositoryState{
			Branch: &git.Branch{
				Name:      "main",
				Pushables: "0",
				Pullables: "1",
				Clean:     false,
			},
		},
	}

	gui := &Gui{}
	header := stripANSI(gui.renderTableHeaderLine(rule))
	_, branchText := align("main", rule.MaxBranch-2, true)
	branch := branchText + " ✗"
	branchPadding, _ := align(branch, rule.MaxBranch, false)
	branch = branch + repeatSpaces(branchPadding)
	repoPadding, repoText := align(repo.Name, rule.MaxName, true)
	row := stripANSI(formatRepositoryTableLine(
		selectionIndicator,
		renderRevCount(repo, rule),
		branch,
		repoText+repeatSpaces(repoPadding),
	))

	if got, want := pipePositions(row), pipePositions(header); len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("separator positions do not match: header=%v row=%v\nheader=%q\nrow=%q", want, got, header, row)
	}

	if got := pipePositions(header); len(got) != 2 {
		t.Fatalf("expected exactly 2 separators in header, got %v in %q", got, header)
	}
}

func TestDisplayWidthIgnoresANSISequences(t *testing.T) {
	if got, want := displayWidth(sep), 3; got != want {
		t.Fatalf("separator width mismatch: got %d want %d", got, want)
	}
}

func pipePositions(in string) []int {
	positions := []int{}
	runeIndex := 0
	for _, r := range in {
		if r == '|' {
			positions = append(positions, runeIndex)
		}
		runeIndex++
	}
	return positions
}

func repeatSpaces(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}
