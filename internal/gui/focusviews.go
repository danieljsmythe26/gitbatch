package gui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/isacikgoz/gitbatch/internal/git"
	"github.com/jroimartin/gocui"
)

const (
	branchOldAgeDays   = 14
	branchStaleAgeDays = 30
)

// listens the event -> "branch.updated"
func (gui *Gui) branchUpdated(event *git.RepositoryEvent) error {
	gui.g.Update(func(g *gocui.Gui) error {
		return gui.renderRepositoryDetails(gui.getSelectedRepository(), true)
	})
	return nil
}

// moves the cursor downwards for the main view and if it goes to bottom it
// prevents from going further
func (gui *Gui) commitCursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, cy := v.Cursor()
		_, oy := v.Origin()
		ly := len(v.BufferLines()) - 1

		// if we are at the end we just return
		if cy+oy == ly-1 {
			return nil
		}
		v.EditDelete(true)
		pos := cy + oy + 1
		_ = adjustAnchor(pos, ly, v)
	}
	return nil
}

// moves the cursor upwards for the main view
func (gui *Gui) commitCursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, oy := v.Origin()
		_, cy := v.Cursor()
		ly := len(v.BufferLines()) - 1
		v.EditDelete(true)
		pos := cy + oy - 1
		_ = adjustAnchor(pos, ly, v)
	}
	return nil
}

// updates the large branch overview for given entity
func (gui *Gui) renderCommits(r *git.Repository) error {
	v, err := gui.g.View(commitViewFeature.Name)
	if err != nil {
		return err
	}
	v.Clear()
	summaries, err := r.BranchSummaries(time.Now())
	if err != nil {
		return err
	}
	v.Title = branchSummaryTitle(summaries)
	width, _ := v.Size()
	si := 0
	for i, summary := range summaries {
		if summary.Current {
			si = i
		}
		fmt.Fprintln(v, gui.branchSummaryLabel(summary, width))
	}
	_ = adjustAnchor(si, len(summaries), v)
	return nil
}

func branchSummaryTitle(summaries []*git.BranchSummary) string {
	worktrees := 0
	stale := 0
	deletable := 0
	for _, summary := range summaries {
		if summary.Worktree {
			worktrees++
		}
		if !summary.Current && summary.AgeDays >= branchStaleAgeDays {
			stale++
		}
		if summary.Merged || summary.UpstreamGone {
			deletable++
		}
	}
	return fmt.Sprintf(" Branches (%d) Worktrees (%d) Stale (%d) Delete (%d) ", len(summaries), worktrees, stale, deletable)
}

func (gui *Gui) branchSummaryLabel(summary *git.BranchSummary, width int) string {
	indicator := unselectedIndicator
	branchColor := cyan
	if summary.Current {
		indicator = green.Sprint(selectionIndicator)
		branchColor = green
	}

	status := gui.branchSummaryStatus(summary)
	statusWidth := displayWidth(stripANSI(status))
	nameWidth := width - displayWidth(indicator) - statusWidth - 2
	if nameWidth < minBranchColumnWidth {
		nameWidth = minBranchColumnWidth
	}
	n, name := align(summary.Name, nameWidth, true)
	return indicator + branchColor.Sprint(name) + strings.Repeat(" ", n+2) + status
}

func (gui *Gui) branchSummaryStatus(summary *git.BranchSummary) string {
	labels := make([]string, 0)
	if summary.Current {
		labels = append(labels, green.Sprint("* current"))
	}
	if summary.Dirty {
		labels = append(labels, yellow.Sprint("dirty"))
	}
	if summary.Worktree {
		labels = append(labels, yellow.Sprint("wt"))
	}
	if summary.UpstreamGone {
		labels = append(labels, red.Sprint("gone del"))
	} else if summary.Merged {
		labels = append(labels, red.Sprint("merged del"))
	}
	if summary.Ahead > 0 {
		labels = append(labels, blue.Sprint("+"+strconv.Itoa(summary.Ahead)))
	}
	if summary.Behind > 0 {
		labels = append(labels, yellow.Sprint("-"+strconv.Itoa(summary.Behind)))
	}
	if !summary.Current && summary.AgeDays >= branchStaleAgeDays {
		labels = append(labels, yellow.Sprint("stale"))
	} else if !summary.Current && summary.AgeDays >= branchOldAgeDays {
		labels = append(labels, yellow.Sprint("old"))
	}
	if len(labels) == 0 {
		return cyan.Sprint("clean")
	}
	return strings.Join(labels, ws)
}

// moves cursor down for a page size
func (gui *Gui) commitPageDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, oy := v.Origin()
		_, vy := v.Size()
		_, cy := v.Cursor()
		lr := len(v.BufferLines())
		if lr < vy {
			return nil
		}
		v.EditDelete(true)

		_ = adjustAnchor(oy+cy+vy-1, lr, v)
		// if err := gui.commitStats(oy + cy + vy - 1); err != nil {
		// 	return err
		// }

	}
	return nil
}

// moves cursor to the top
func (gui *Gui) commitCursorTop(g *gocui.Gui, v *gocui.View) error {
	if v != nil {

		v.EditDelete(true)
		lr := len(v.BufferLines())

		_ = adjustAnchor(0, lr, v)
	}
	return nil
}

// moves cursor up for a page size
func (gui *Gui) commitPageUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, oy := v.Origin()
		_, cy := v.Cursor()
		_, vy := v.Size()
		lr := len(v.BufferLines())
		v.EditDelete(true)
		_ = adjustAnchor(oy+cy-vy+1, lr, v)
		// if err := gui.commitStats(oy + cy - vy + 1); err != nil {
		// 	return err
		// }
	}
	return nil
}
