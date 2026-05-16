package gui

import (
	"sync"
	"testing"
	"time"

	"github.com/isacikgoz/gitbatch/internal/git"
	"github.com/isacikgoz/gitbatch/internal/job"
	"github.com/jroimartin/gocui"
)

func TestNewPreservesCheckoutMode(t *testing.T) {
	gui, err := New(string(CheckoutMode), nil, nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got := gui.State.Mode.ModeID; got != CheckoutMode {
		t.Fatalf("expected mode %q, got %q", CheckoutMode, got)
	}
}

func TestJobsForPrimaryActionPushesCurrentRepoWhenQueueEmpty(t *testing.T) {
	gui, err := New(string(PushMode), nil, nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	repo := &git.Repository{
		RepoID: "repo-1",
		State: &git.RepositoryState{
			Branch: &git.Branch{
				Name:      "main",
				Pushables: "3",
				Pullables: "0",
				Upstream:  &git.RemoteBranch{Name: "origin/main"},
			},
			Remote: &git.Remote{Name: "origin"},
		},
	}

	jobs, err := gui.jobsForPrimaryAction(repo)
	if err != nil {
		t.Fatalf("jobsForPrimaryAction returned error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 direct push job, got %d", len(jobs))
	}
	if jobs[0].JobType != job.PushJob {
		t.Fatalf("expected push job, got %q", jobs[0].JobType)
	}
	if jobs[0].Repository != repo {
		t.Fatalf("expected selected repo to be pushed")
	}
}

func TestPushFeedbackExpires(t *testing.T) {
	previousDuration := pushFeedbackDuration
	pushFeedbackDuration = 5 * time.Millisecond
	defer func() {
		pushFeedbackDuration = previousDuration
	}()

	gui := &Gui{
		State: guiState{
			pushFeedback: make(map[string]pushFeedbackState),
		},
		feedbackMu: &sync.RWMutex{},
	}

	repo := &git.Repository{RepoID: "repo-1"}
	gui.setPushFeedback(repo, true, "pushed to origin")

	if feedback, ok := gui.pushFeedbackFor(repo); !ok || !feedback.Success {
		t.Fatalf("expected push feedback to be visible immediately")
	}

	time.Sleep(20 * time.Millisecond)

	if _, ok := gui.pushFeedbackFor(repo); ok {
		t.Fatalf("expected push feedback to expire")
	}
}

func TestStatusModeBindingsIncludeRefreshAndResetAll(t *testing.T) {
	gui, err := New(string(FetchMode), nil, nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	bindings := gui.dynamicModeBindings(StatusMode)

	var hasRefresh bool
	var hasResetAll bool
	for _, binding := range bindings {
		switch binding.Key {
		case 'r':
			if binding.Display == "r" && binding.Description == "refresh" && binding.Vital {
				hasRefresh = true
			}
		case gocui.KeyCtrlR:
			if binding.Display == "c-r" && binding.Description == "reset all" && binding.Vital {
				hasResetAll = true
			}
		}
	}

	if !hasRefresh {
		t.Fatalf("expected status mode to include plain r refresh binding")
	}
	if !hasResetAll {
		t.Fatalf("expected status mode to retain ctrl-r reset all binding")
	}
}

func TestMainBindingsIncludeRefresh(t *testing.T) {
	gui, err := New(string(FetchMode), nil, nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if err := gui.generateKeybindings(); err != nil {
		t.Fatalf("generateKeybindings returned error: %v", err)
	}

	for _, binding := range gui.KeyBindings {
		if binding.View == mainViewFeature.Name && binding.Key == 'r' && binding.Description == "Refresh" && binding.Vital {
			return
		}
	}
	t.Fatalf("expected main view to include r Refresh binding")
}

func TestSmartRowsGroupPinnedAttentionRecentAndQuiet(t *testing.T) {
	now := time.Now()
	pinned := testRepo("ai-estimates-app", now.Add(-48*time.Hour), true, "0", "0")
	attention := testRepo("expenses", now.Add(-72*time.Hour), false, "0", "0")
	recent := testRepo("gitbatch", now.Add(-1*time.Hour), true, "0", "0")
	quiet := testRepo("agent-config", now.Add(-14*24*time.Hour), true, "0", "0")

	gui := &Gui{
		State: guiState{
			Repositories:       []*git.Repository{quiet, recent, attention, pinned},
			PinnedRepositories: []string{"ai-estimates-app"},
			SortMode:           repositorySortSmart,
		},
	}

	gui.sortRepositoriesSmart()
	rows := gui.mainRows()

	got := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Repository != nil {
			got = append(got, row.Repository.Name)
			continue
		}
		got = append(got, row.Label)
	}
	want := []string{
		"Pinned",
		"ai-estimates-app",
		"Needs attention",
		"expenses",
		"Recent",
		"gitbatch",
		"Quiet",
		"agent-config",
	}
	if len(got) != len(want) {
		t.Fatalf("smart rows length mismatch:\ngot  %v\nwant %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("smart rows mismatch:\ngot  %v\nwant %v", got, want)
		}
	}
}

func testRepo(name string, modTime time.Time, clean bool, pushables string, pullables string) *git.Repository {
	return &git.Repository{
		Name:    name,
		ModTime: modTime,
		State: &git.RepositoryState{
			Branch: &git.Branch{
				Name:      "main",
				Clean:     clean,
				Pushables: pushables,
				Pullables: pullables,
			},
		},
	}
}
