package gui

import (
	"github.com/isacikgoz/gitbatch/internal/git"
	"github.com/isacikgoz/gitbatch/internal/job"
)

func (gui *Gui) queueLen() int {
	gui.queueMu.RLock()
	defer gui.queueMu.RUnlock()
	return gui.State.Queue.Len()
}

func (gui *Gui) queueJobs() []*job.Job {
	gui.queueMu.RLock()
	defer gui.queueMu.RUnlock()
	return gui.State.Queue.Jobs()
}

func (gui *Gui) queueAddJob(j *job.Job) error {
	gui.queueMu.Lock()
	defer gui.queueMu.Unlock()
	return gui.State.Queue.AddJob(j)
}

func (gui *Gui) queueRemoveJob(r *git.Repository) error {
	gui.queueMu.Lock()
	defer gui.queueMu.Unlock()
	return gui.State.Queue.RemoveFromQueue(r)
}

func (gui *Gui) queueIsInTheQueue(r *git.Repository) (bool, *job.Job) {
	gui.queueMu.RLock()
	defer gui.queueMu.RUnlock()
	return gui.State.Queue.IsInTheQueue(r)
}

func (gui *Gui) replaceQueue(q *job.Queue) {
	gui.queueMu.Lock()
	defer gui.queueMu.Unlock()
	gui.State.Queue = q
}

func (gui *Gui) failoverAddJob(j *job.Job) error {
	gui.queueMu.Lock()
	defer gui.queueMu.Unlock()
	return gui.State.FailoverQueue.AddJob(j)
}

func (gui *Gui) failoverRemoveJob(r *git.Repository) error {
	gui.queueMu.Lock()
	defer gui.queueMu.Unlock()
	return gui.State.FailoverQueue.RemoveFromQueue(r)
}

func (gui *Gui) failoverIsInTheQueue(r *git.Repository) (bool, *job.Job) {
	gui.queueMu.RLock()
	defer gui.queueMu.RUnlock()
	return gui.State.FailoverQueue.IsInTheQueue(r)
}
