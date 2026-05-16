package gui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/isacikgoz/gitbatch/internal/command"
	gerr "github.com/isacikgoz/gitbatch/internal/errors"
	"github.com/isacikgoz/gitbatch/internal/git"
	"github.com/isacikgoz/gitbatch/internal/job"
	"github.com/jroimartin/gocui"
)

// refresh the main view and re-render the repository representations
func (gui *Gui) renderMain() error {
	gui.mutex.Lock()
	defer gui.mutex.Unlock()

	mainView, err := gui.g.View(mainViewFeature.Name)
	if err != nil {
		return err
	}
	mainView.Clear()
	gui.State.mainRows = gui.mainRows()
	if err := gui.ensureMainCursorOnRepository(mainView); err != nil {
		return err
	}
	rules := gui.renderRules()
	gui.renderTableHeader(rules)
	for _, row := range gui.State.mainRows {
		if row.Repository == nil {
			fmt.Fprintln(mainView, gui.renderMainGroupHeader(row.Label, rules))
			continue
		}
		fmt.Fprintln(mainView, gui.repositoryLabelWithRules(row.Repository, rules))
	}
	selected := gui.getSelectedRepository()
	selectionChanged := selected != nil && gui.State.detailRepoID != selected.RepoID
	if selected != nil {
		gui.State.detailRepoID = selected.RepoID
	}
	return gui.renderRepositoryDetails(selected, selectionChanged)
}

// listens the event -> "repository.updated"
func (gui *Gui) repositoryUpdated(event *git.RepositoryEvent) error {
	gui.g.Update(func(g *gocui.Gui) error {
		return gui.renderMain()
	})
	return nil
}

func (gui *Gui) renderRepositoryDetails(r *git.Repository, resetDynamic bool) error {
	if r == nil {
		return nil
	}
	_ = r.State.Branch.InitializeCommits(r)
	if err := gui.renderSideViews(r); err != nil {
		return err
	}
	if err := gui.renderCommits(r); err != nil {
		return err
	}
	if err := gui.initStashedView(r); err != nil {
		return err
	}
	if resetDynamic || gui.currentDynamicMode() == StatusMode {
		return gui.initFocusStat(r)
	}
	return nil
}

func (gui *Gui) mainRows() []mainRow {
	if gui.State.SortMode != repositorySortSmart {
		rows := make([]mainRow, 0, len(gui.State.Repositories))
		for _, r := range gui.State.Repositories {
			rows = append(rows, mainRow{Repository: r})
		}
		return rows
	}

	groups := []struct {
		label string
		rows  []mainRow
	}{
		{label: "Pinned"},
		{label: "Needs attention"},
		{label: "Recent"},
		{label: "Quiet"},
	}
	for _, r := range gui.State.Repositories {
		index := gui.repositoryGroupIndex(r)
		groups[index].rows = append(groups[index].rows, mainRow{Repository: r})
	}

	rows := make([]mainRow, 0, len(gui.State.Repositories)+len(groups))
	for _, group := range groups {
		if len(group.rows) == 0 {
			continue
		}
		rows = append(rows, mainRow{Label: group.label})
		rows = append(rows, group.rows...)
	}
	return rows
}

func (gui *Gui) renderMainGroupHeader(label string, rule *RepositoryDecorationRules) string {
	width := displayWidth(gui.renderTableHeaderLine(rule))
	text := "── " + label + " "
	if width > displayWidth(text) {
		text += strings.Repeat("─", width-displayWidth(text))
	}
	return yellow.Sprint(text)
}

func (gui *Gui) sortRepositoriesSmart() {
	sort.SliceStable(gui.State.Repositories, func(i, j int) bool {
		return gui.lessRepository(gui.State.Repositories[i], gui.State.Repositories[j])
	})
}

func (gui *Gui) lessRepository(left, right *git.Repository) bool {
	leftGroup := gui.repositoryGroupIndex(left)
	rightGroup := gui.repositoryGroupIndex(right)
	if leftGroup != rightGroup {
		return leftGroup < rightGroup
	}
	leftPin, leftPinned := gui.pinnedRepositoryRank(left)
	rightPin, rightPinned := gui.pinnedRepositoryRank(right)
	if leftPinned && rightPinned && leftPin != rightPin {
		return leftPin < rightPin
	}
	if left.ModTime.Unix() != right.ModTime.Unix() {
		return left.ModTime.Unix() > right.ModTime.Unix()
	}
	return git.Less(left, right)
}

func (gui *Gui) repositoryGroupIndex(r *git.Repository) int {
	if _, ok := gui.pinnedRepositoryRank(r); ok {
		return 0
	}
	if repositoryNeedsAttention(r) {
		return 1
	}
	if time.Since(r.ModTime) <= 7*24*time.Hour {
		return 2
	}
	return 3
}

func (gui *Gui) pinnedRepositoryRank(r *git.Repository) (int, bool) {
	if r == nil {
		return 0, false
	}
	for i, name := range gui.State.PinnedRepositories {
		if strings.EqualFold(r.Name, name) {
			return i, true
		}
	}
	return 0, false
}

func repositoryNeedsAttention(r *git.Repository) bool {
	if r == nil || r.State == nil {
		return false
	}
	status := r.WorkStatus()
	if status == git.Queued || status == git.Working || status == git.Paused || status == git.Fail {
		return true
	}
	if r.State.Branch == nil {
		return false
	}
	if !r.State.Branch.Clean {
		return true
	}
	return revNeedsAttention(r.State.Branch.Pushables) || revNeedsAttention(r.State.Branch.Pullables)
}

func revNeedsAttention(value string) bool {
	return value != "" && value != "0"
}

// moves the cursor downwards for the main view and if it goes to bottom it
// prevents from going further
func (gui *Gui) cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, cy := v.Cursor()
		_, oy := v.Origin()
		next, ok := gui.nextRepositoryRow(oy+cy, 1)
		if !ok {
			return nil
		}
		if err := gui.setMainCursorToRow(v, next); err != nil {
			return err
		}
	}
	return gui.renderMain()
}

func (gui *Gui) ensureMainCursorOnRepository(v *gocui.View) error {
	if v == nil {
		return nil
	}
	_, oy := v.Origin()
	_, cy := v.Cursor()
	row := oy + cy
	if row >= 0 && row < len(gui.State.mainRows) && gui.State.mainRows[row].Repository != nil {
		return nil
	}
	if next, ok := gui.repositoryRowAtOrAfter(row); ok {
		return gui.setMainCursorToRow(v, next)
	}
	if prev, ok := gui.repositoryRowAtOrBefore(row); ok {
		return gui.setMainCursorToRow(v, prev)
	}
	return nil
}

func (gui *Gui) setMainCursorToRow(v *gocui.View, row int) error {
	if v == nil {
		return nil
	}
	if row < 0 {
		row = 0
	}
	if max := len(gui.State.mainRows) - 1; row > max {
		row = max
	}
	cx, _ := v.Cursor()
	ox, oy := v.Origin()
	_, height := v.Size()
	if height <= 0 {
		return nil
	}
	if row < oy {
		if err := v.SetOrigin(ox, row); err != nil {
			return err
		}
		return v.SetCursor(cx, 0)
	}
	if row >= oy+height {
		nextOrigin := row - height + 1
		if err := v.SetOrigin(ox, nextOrigin); err != nil {
			return err
		}
		return v.SetCursor(cx, height-1)
	}
	return v.SetCursor(cx, row-oy)
}

func (gui *Gui) nextRepositoryRow(start int, direction int) (int, bool) {
	if direction == 0 {
		return start, false
	}
	for row := start + direction; row >= 0 && row < len(gui.State.mainRows); row += direction {
		if gui.State.mainRows[row].Repository != nil {
			return row, true
		}
	}
	return 0, false
}

func (gui *Gui) repositoryRowAtOrAfter(start int) (int, bool) {
	if start < 0 {
		start = 0
	}
	for row := start; row < len(gui.State.mainRows); row++ {
		if gui.State.mainRows[row].Repository != nil {
			return row, true
		}
	}
	return 0, false
}

func (gui *Gui) repositoryRowAtOrBefore(start int) (int, bool) {
	if start >= len(gui.State.mainRows) {
		start = len(gui.State.mainRows) - 1
	}
	for row := start; row >= 0; row-- {
		if gui.State.mainRows[row].Repository != nil {
			return row, true
		}
	}
	return 0, false
}

func (gui *Gui) firstRepositoryRow() (int, bool) {
	return gui.repositoryRowAtOrAfter(0)
}

func (gui *Gui) lastRepositoryRow() (int, bool) {
	return gui.repositoryRowAtOrBefore(len(gui.State.mainRows) - 1)
}

// moves the cursor upwards for the main view
func (gui *Gui) cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, oy := v.Origin()
		_, cy := v.Cursor()
		prev, ok := gui.nextRepositoryRow(oy+cy, -1)
		if !ok {
			return nil
		}
		if err := gui.setMainCursorToRow(v, prev); err != nil {
			return err
		}
	}
	return gui.renderMain()
}

// moves cursor to the top
func (gui *Gui) cursorTop(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		if first, ok := gui.firstRepositoryRow(); ok {
			if err := gui.setMainCursorToRow(v, first); err != nil {
				return err
			}
		}
	}
	return gui.renderMain()
}

// moves cursor to the end
func (gui *Gui) cursorEnd(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		if last, ok := gui.lastRepositoryRow(); ok {
			if err := gui.setMainCursorToRow(v, last); err != nil {
				return err
			}
		}
	}
	return gui.renderMain()
}

// moves cursor down for a page size
func (gui *Gui) pageDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, oy := v.Origin()
		_, cy := v.Cursor()
		_, vy := v.Size()
		if len(gui.State.mainRows) < vy {
			return nil
		}
		target := oy + cy + vy
		if row, ok := gui.repositoryRowAtOrAfter(target); ok {
			if err := gui.setMainCursorToRow(v, row); err != nil {
				return err
			}
			return gui.renderMain()
		}
		if row, ok := gui.lastRepositoryRow(); ok {
			if err := gui.setMainCursorToRow(v, row); err != nil {
				return err
			}
		}
	}
	return gui.renderMain()
}

// moves cursor up for a page size
func (gui *Gui) pageUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, oy := v.Origin()
		_, cy := v.Cursor()
		_, vy := v.Size()
		target := oy + cy - vy
		if target < 0 {
			target = 0
		}
		if row, ok := gui.repositoryRowAtOrBefore(target); ok {
			if err := gui.setMainCursorToRow(v, row); err != nil {
				return err
			}
			return gui.renderMain()
		}
		if row, ok := gui.firstRepositoryRow(); ok {
			if err := gui.setMainCursorToRow(v, row); err != nil {
				return err
			}
		}
	}
	return gui.renderMain()
}

// returns the repository at the cursor position, skipping display-only group rows.
func (gui *Gui) getSelectedRepository() *git.Repository {
	if len(gui.State.Repositories) == 0 {
		return nil
	}
	if gui.g == nil {
		return nil
	}
	v, err := gui.g.View(mainViewFeature.Name)
	if err != nil {
		return nil
	}
	_, oy := v.Origin()
	_, cy := v.Cursor()
	row := oy + cy
	if row < 0 || row >= len(gui.State.mainRows) {
		return nil
	}
	return gui.State.mainRows[row].Repository
}

// adds given entity to job queue
func (gui *Gui) addToQueue(r *git.Repository) error {
	j, err := gui.newJobForRepository(r)
	if err != nil || j == nil {
		return err
	}
	err = gui.queueAddJob(j)
	if err != nil {
		return err
	}
	r.SetWorkStatus(git.Queued)
	return nil
}

func (gui *Gui) newJobForRepository(r *git.Repository) (*job.Job, error) {
	if r == nil {
		return nil, nil
	}

	j := &job.Job{
		Repository: r,
	}
	switch mode := gui.State.Mode.ModeID; mode {
	case FetchMode:
		j.JobType = job.FetchJob
	case PullMode:
		if r.State.Branch.Upstream == nil {
			return nil, nil
		}
		j.JobType = job.PullJob
	case PushMode:
		if r.State.Branch.Upstream == nil {
			gui.setPushFeedback(r, false, "Push unavailable: branch is not tracking a remote branch.")
			return nil, nil
		}
		j.JobType = job.PushJob
	case MergeMode:
		if r.State.Branch.Upstream == nil {
			return nil, nil
		}
		j.JobType = job.MergeJob
	case CheckoutMode:
		j.JobType = job.CheckoutJob
		j.Options = &command.CheckoutOptions{
			TargetRef:      gui.State.targetBranch,
			CreateIfAbsent: true,
		}
	default:
		return nil, nil
	}
	return j, nil
}

// removes given entity from job queue
func (gui *Gui) removeFromQueue(r *git.Repository) error {
	err := gui.queueRemoveJob(r)
	if err != nil {
		return err
	}
	r.SetWorkStatus(git.Available)
	return nil
}

// this function starts the queue and updates the gui with the result of an
// operation
func (gui *Gui) startQueue(g *gocui.Gui, v *gocui.View) error {
	jobs := gui.queueJobs()
	if len(jobs) == 0 {
		return nil
	}
	gui.runJobs(jobs)
	return nil
}

func (gui *Gui) startPrimaryAction(g *gocui.Gui, v *gocui.View) error {
	jobs, err := gui.jobsForPrimaryAction(gui.getSelectedRepository())
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return nil
	}
	gui.runJobs(jobs)
	return nil
}

func (gui *Gui) jobsForPrimaryAction(selected *git.Repository) ([]*job.Job, error) {
	if gui.State.Mode.ModeID != PushMode || gui.queueLen() > 0 {
		return gui.queueJobs(), nil
	}

	j, err := gui.newJobForRepository(selected)
	if err != nil || j == nil {
		return nil, err
	}
	return []*job.Job{j}, nil
}

func (gui *Gui) runJobs(jobs []*job.Job) {
	go func(guiGo *Gui, jobsToRun []*job.Job) {
		queue := job.CreateJobQueue()
		for _, queuedJob := range jobsToRun {
			if err := queue.AddJob(queuedJob); err != nil {
				continue
			}
		}

		fails := queue.StartJobsAsync()
		guiGo.replaceQueue(job.CreateJobQueue())
		for _, queuedJob := range jobsToRun {
			err, failed := fails[queuedJob]
			if failed {
				if err == gerr.ErrAuthenticationRequired {
					queuedJob.Repository.SetWorkStatus(git.Paused)
					_ = guiGo.failoverAddJob(queuedJob)
					if queuedJob.JobType == job.PushJob {
						guiGo.setPushFeedback(queuedJob.Repository, false, "Push paused: authentication required. Press [u].")
					}
					continue
				}
				if queuedJob.JobType == job.PushJob {
					guiGo.setPushFeedback(queuedJob.Repository, false, "Push failed: "+queuedJob.Repository.State.Message)
				}
				continue
			}
			if queuedJob.JobType == job.PushJob && queuedJob.Repository.WorkStatus() == git.Success {
				guiGo.setPushFeedback(queuedJob.Repository, true, queuedJob.Repository.State.Message)
			}
		}
		guiGo.requestRender()
	}(gui, jobs)
}

func (gui *Gui) submitCredentials(g *gocui.Gui, v *gocui.View) error {
	if is, j := gui.failoverIsInTheQueue(gui.getSelectedRepository()); is {
		if j.Repository.WorkStatus() == git.Paused {
			if err := gui.failoverRemoveJob(j.Repository); err != nil {
				return err
			}
			err := gui.openAuthenticationView(g, nil, j, v.Name())
			if err != nil {
				return err
			}
			if isnt, _ := gui.queueIsInTheQueue(j.Repository); !isnt {
				_ = gui.failoverAddJob(j)
			}
		}
	}
	return nil
}

// marking repository is simply adding the repository into the queue. the
// function does take its current state into account before adding it
func (gui *Gui) markRepository(g *gocui.Gui, v *gocui.View) error {
	r := gui.getSelectedRepository()
	// maybe, failed entities may be added to queue again
	if r == nil {
		return nil
	}
	if r.WorkStatus().Ready {
		if err := gui.addToQueue(r); err != nil {
			return err
		}
	} else if r.WorkStatus() == git.Queued {
		if err := gui.removeFromQueue(r); err != nil {
			return err
		}
	}
	return nil
}

// add all remaining repositories into the queue. the function does take its
// current state into account before adding it
func (gui *Gui) markAllRepositories(g *gocui.Gui, v *gocui.View) error {
	for _, r := range gui.State.Repositories {
		if r.WorkStatus().Ready {
			if err := gui.addToQueue(r); err != nil {
				return err
			}
		} else {
			continue
		}
	}
	return nil
}

// remove all repositories from the queue. the function does take its
// current state into account before removing it
func (gui *Gui) unmarkAllRepositories(g *gocui.Gui, v *gocui.View) error {
	for _, r := range gui.State.Repositories {
		if r.WorkStatus() == git.Queued {
			if err := gui.removeFromQueue(r); err != nil {
				return err
			}
		} else {
			continue
		}
	}
	return nil
}

// sortByName sorts the repositories by A to Z order
func (gui *Gui) sortByName(g *gocui.Gui, v *gocui.View) error {
	gui.State.SortMode = repositorySortName
	sort.Sort(git.Alphabetical(gui.State.Repositories))
	_ = gui.renderMain()
	_ = gui.renderTitle()
	return nil
}

// sortByMod sorts the repositories according to last modifed date
// the top element will be the last modified
func (gui *Gui) sortByMod(g *gocui.Gui, v *gocui.View) error {
	gui.State.SortMode = repositorySortDate
	sort.Sort(git.LastModified(gui.State.Repositories))
	_ = gui.renderMain()
	_ = gui.renderTitle()
	return nil
}

// sortBySmart sorts repositories into pinned, attention, recent, then quiet buckets.
func (gui *Gui) sortBySmart(g *gocui.Gui, v *gocui.View) error {
	gui.State.SortMode = repositorySortSmart
	gui.sortRepositoriesSmart()
	_ = gui.renderMain()
	_ = gui.renderTitle()
	return nil
}
