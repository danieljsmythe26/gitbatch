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
	gui, err := New(string(CheckoutMode), nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got := gui.State.Mode.ModeID; got != CheckoutMode {
		t.Fatalf("expected mode %q, got %q", CheckoutMode, got)
	}
}

func TestJobsForPrimaryActionPushesCurrentRepoWhenQueueEmpty(t *testing.T) {
	gui, err := New(string(PushMode), nil)
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
	gui, err := New(string(FetchMode), nil)
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
