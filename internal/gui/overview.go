package gui

import (
	"github.com/jroimartin/gocui"
)

// set the layout and create views with their default size, name etc. values
// TODO: window sizes can be handled better
func (gui *Gui) overviewLayout(g *gocui.Gui) error {
	g.SelFgColor = gocui.ColorGreen
	maxX, maxY := g.Size()
	leftContentWidth := int(0.32 * float32(maxX))
	if leftContentWidth < 50 {
		leftContentWidth = 50
	}
	leftWidth := leftContentWidth + 2
	minPaneWidth := 2

	rightWidth := int(0.18 * float32(maxX))
	if rightWidth < minPaneWidth {
		rightWidth = minPaneWidth
	}

	commitMinWidth := 20
	statusMinWidth := 20
	rightMinWidth := 28
	middleMinWidth := commitMinWidth + statusMinWidth
	maxLeftWidth := maxX - rightMinWidth - middleMinWidth
	if maxLeftWidth < 30 {
		maxLeftWidth = 30
	}
	if leftWidth > maxLeftWidth {
		leftWidth = maxLeftWidth
	}
	if maxX-leftWidth < minPaneWidth*3 {
		leftWidth = maxX - (minPaneWidth * 3)
	}
	if leftWidth < 10 {
		leftWidth = 10
	}

	detailRight := maxX - rightWidth
	maxRightWidth := maxX - leftWidth - (minPaneWidth * 2)
	if maxRightWidth < minPaneWidth {
		maxRightWidth = minPaneWidth
	}
	if rightWidth > maxRightWidth {
		rightWidth = maxRightWidth
	}
	if detailRight < leftWidth+(minPaneWidth*2) {
		detailRight = leftWidth + (minPaneWidth * 2)
	}
	if detailRight > maxX-minPaneWidth {
		detailRight = maxX - minPaneWidth
	}

	detailMiddle := leftWidth + (detailRight-leftWidth)/2
	if detailMiddle-leftWidth < minPaneWidth {
		detailMiddle = leftWidth + minPaneWidth
	}
	if detailRight-detailMiddle < minPaneWidth {
		detailMiddle = detailRight - minPaneWidth
	}
	if detailMiddle <= leftWidth {
		detailMiddle = leftWidth + minPaneWidth
	}
	if detailRight <= detailMiddle {
		detailRight = detailMiddle + minPaneWidth
	}
	if detailRight > maxX-1 {
		detailRight = maxX - 1
	}
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
	if v, err := g.SetView(commitViewFeature.Name, leftWidth, 0, detailMiddle-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = commitViewFeature.Title
		v.Wrap = false
		v.Autoscroll = false
	}
	if v, err := g.SetView(dynamicViewFeature.Name, detailMiddle, 0, detailRight-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = dynamicViewFeature.Title
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

// close confirmation view
func (gui *Gui) openBranchesView(g *gocui.Gui, v *gocui.View) error {
	return gui.focusToView(branchViewFeature.Name)
}

// close confirmation view
func (gui *Gui) closeBranchesView(g *gocui.Gui, v *gocui.View) error {
	return gui.focusToView(mainViewFeature.Name)
}
