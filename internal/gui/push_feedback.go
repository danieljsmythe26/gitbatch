package gui

import (
	"sync"
	"time"

	"github.com/isacikgoz/gitbatch/internal/git"
	"github.com/jroimartin/gocui"
)

var pushFeedbackDuration = 1500 * time.Millisecond

func (gui *Gui) setPushFeedback(r *git.Repository, success bool, message string) {
	if r == nil {
		return
	}
	if gui.feedbackMu == nil {
		gui.feedbackMu = &sync.RWMutex{}
	}
	expiresAt := time.Now().Add(pushFeedbackDuration)

	gui.feedbackMu.Lock()
	if gui.State.pushFeedback == nil {
		gui.State.pushFeedback = make(map[string]pushFeedbackState)
	}
	gui.State.pushFeedback[r.RepoID] = pushFeedbackState{
		Message:   message,
		Success:   success,
		ExpiresAt: expiresAt,
	}
	gui.feedbackMu.Unlock()

	gui.requestRender()

	go func(repoID string, expiry time.Time) {
		time.Sleep(pushFeedbackDuration)

		shouldRender := false
		gui.feedbackMu.Lock()
		feedback, ok := gui.State.pushFeedback[repoID]
		if ok && feedback.ExpiresAt.Equal(expiry) {
			delete(gui.State.pushFeedback, repoID)
			shouldRender = true
		}
		gui.feedbackMu.Unlock()

		if shouldRender {
			gui.requestRender()
		}
	}(r.RepoID, expiresAt)
}

func (gui *Gui) pushFeedbackFor(r *git.Repository) (pushFeedbackState, bool) {
	if r == nil {
		return pushFeedbackState{}, false
	}
	if gui.feedbackMu == nil {
		return pushFeedbackState{}, false
	}

	gui.feedbackMu.RLock()
	feedback, ok := gui.State.pushFeedback[r.RepoID]
	gui.feedbackMu.RUnlock()
	if !ok {
		return pushFeedbackState{}, false
	}
	if time.Now().After(feedback.ExpiresAt) {
		return pushFeedbackState{}, false
	}
	return feedback, true
}

func (gui *Gui) hasSuccessfulPushFeedback(r *git.Repository) bool {
	feedback, ok := gui.pushFeedbackFor(r)
	return ok && feedback.Success
}

func (gui *Gui) requestRender() {
	if gui.g == nil {
		return
	}
	gui.g.Update(func(g *gocui.Gui) error {
		return gui.renderMain()
	})
}
