package gui

import (
	"strings"

	"github.com/jroimartin/gocui"
)

const (
	leftColumnPercent    = 40
	middleColumnPercent  = 40
	statusMaxPercent     = 60
	statusDefaultPercent = 35
	minDetailPaneHeight  = 4
	minCommitPaneHeight  = 8
)

// set the layout and create views with their default size, name etc. values
// TODO: window sizes can be handled better
func (gui *Gui) overviewLayout(g *gocui.Gui) error {
	g.SelFgColor = gocui.ColorGreen
	maxX, maxY := g.Size()
	leftWidth, middleWidth := gui.overviewColumnWidths(maxX)
	detailRight := leftWidth + middleWidth
	statusBottom := gui.statusPaneBottom(maxY)
	quarterY := int(0.25 * float32(maxY))
	threeQuarterY := int(0.75 * float32(maxY))

	if v, err := g.SetView(mainViewFrameFeature.Name, 0, 0, leftWidth-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = mainViewFrameFeature.Title
	}
	if v, err := g.SetView(mainViewFeature.Name, 1, 2, leftWidth-2, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = mainViewFeature.Title
		v.Overwrite = true
		if _, err := g.SetCurrentView(mainViewFeature.Name); err != nil {
			return err
		}
		v.Frame = false
	}
	if v, err := g.SetView(dynamicViewFeature.Name, leftWidth, 0, detailRight-1, statusBottom-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = string(StatusMode)
		v.Wrap = false
		v.Autoscroll = false
	}
	if v, err := g.SetView(commitViewFeature.Name, leftWidth, statusBottom, detailRight-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = commitViewFeature.Title
		v.Wrap = false
		v.Autoscroll = false
	}
	if v, err := g.SetView(remoteViewFeature.Name, detailRight, 0, maxX-1, quarterY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = remoteViewFeature.Title
		v.Wrap = false
		v.Autoscroll = false
	}
	if v, err := g.SetView(branchViewFeature.Name, detailRight, quarterY, maxX-1, threeQuarterY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = branchViewFeature.Title
		v.Wrap = false
		v.Autoscroll = false
	}
	if v, err := g.SetView(stashViewFeature.Name, detailRight, threeQuarterY, maxX-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = stashViewFeature.Title
		v.Wrap = false
		v.Autoscroll = false
	}
	if v, err := g.SetView(remoteBranchViewFeature.Name, int(0.25*float32(maxX)), int(0.15*float32(maxY)), int(0.75*float32(maxX)), int(0.75*float32(maxY))); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = remoteBranchViewFeature.Title
		v.Wrap = false
		v.Overwrite = false
		_, _ = g.SetViewOnBottom(v.Name())
	}
	if v, err := g.SetView(batchBranchViewFeature.Name, int(0.25*float32(maxX)), int(0.25*float32(maxY)), int(0.75*float32(maxX)), int(0.75*float32(maxY))); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = batchBranchViewFeature.Title
		v.Wrap = false
		v.Autoscroll = false
		_, _ = g.SetViewOnBottom(v.Name())
	}
	if v, err := g.SetView(suggestBranchViewFeature.Name, int(0.30*float32(maxX)), int(0.45*float32(maxY)), int(0.70*float32(maxX)), int(0.55*float32(maxY))); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = suggestBranchViewFeature.Title
		v.Editable = true
		v.Wrap = false
		v.Autoscroll = false
		_, _ = g.SetViewOnBottom(v.Name())
	}
	if v, err := g.SetView(keybindingsViewFeature.Name, -1, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.BgColor = gocui.ColorWhite
		v.FgColor = gocui.ColorBlack
		v.Frame = false
		_ = gui.updateKeyBindingsView(g, mainViewFeature.Name)
	}
	return nil
}

func (gui *Gui) overviewColumnWidths(maxX int) (int, int) {
	leftWidth := maxX * leftColumnPercent / 100
	middleWidth := maxX * middleColumnPercent / 100

	minColumnWidth := 4
	if leftWidth < minColumnWidth {
		leftWidth = minColumnWidth
	}
	if middleWidth < minColumnWidth {
		middleWidth = minColumnWidth
	}
	if maxX-leftWidth-middleWidth < minColumnWidth {
		middleWidth = maxX - leftWidth - minColumnWidth
	}
	if middleWidth < minColumnWidth {
		middleWidth = minColumnWidth
		leftWidth = maxX - middleWidth - minColumnWidth
	}
	if leftWidth < minColumnWidth {
		leftWidth = minColumnWidth
	}
	return leftWidth, middleWidth
}

func (gui *Gui) statusPaneBottom(maxY int) int {
	detailBottom := maxY - 2
	if detailBottom < minDetailPaneHeight*2 {
		return detailBottom / 2
	}

	minStatusHeight := minDetailPaneHeight
	maxStatusHeight := detailBottom * statusMaxPercent / 100
	if maxStatusHeight < minStatusHeight {
		maxStatusHeight = minStatusHeight
	}
	if detailBottom-minCommitPaneHeight < maxStatusHeight {
		maxStatusHeight = detailBottom - minCommitPaneHeight
	}
	if maxStatusHeight < minStatusHeight {
		maxStatusHeight = minStatusHeight
	}

	statusHeight := detailBottom * statusDefaultPercent / 100
	if dynamicView, err := gui.g.View(dynamicViewFeature.Name); err == nil {
		if lineCount := visibleBufferLineCount(dynamicView.BufferLines()); lineCount > 0 {
			statusHeight = lineCount + 3
		}
	}

	if statusHeight < minStatusHeight {
		statusHeight = minStatusHeight
	}
	if statusHeight > maxStatusHeight {
		statusHeight = maxStatusHeight
	}
	return statusHeight
}

func visibleBufferLineCount(lines []string) int {
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return i + 1
		}
	}
	return 0
}

func (gui *Gui) reflowMiddleColumn() error {
	if gui.g == nil {
		return nil
	}
	maxX, maxY := gui.g.Size()
	leftWidth, middleWidth := gui.overviewColumnWidths(maxX)
	detailRight := leftWidth + middleWidth
	statusBottom := gui.statusPaneBottom(maxY)

	if _, err := gui.g.SetView(dynamicViewFeature.Name, leftWidth, 0, detailRight-1, statusBottom-1); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	if _, err := gui.g.SetView(commitViewFeature.Name, leftWidth, statusBottom, detailRight-1, maxY-2); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	return nil
}

// close confirmation view
func (gui *Gui) openBranchesView(g *gocui.Gui, v *gocui.View) error {
	return gui.focusToView(branchViewFeature.Name)
}

// close confirmation view
func (gui *Gui) closeBranchesView(g *gocui.Gui, v *gocui.View) error {
	return gui.focusToView(mainViewFeature.Name)
}
