package gui

import "github.com/jroimartin/gocui"

var (
	focusViews = []viewFeature{commitViewFeature, dynamicViewFeature, remoteViewFeature, branchViewFeature, stashViewFeature}
	paneViews  = []viewFeature{
		mainViewFeature,
		commitViewFeature,
		dynamicViewFeature,
		remoteViewFeature,
		branchViewFeature,
		stashViewFeature,
	}
	layoutViews = []viewFeature{
		mainViewFrameFeature,
		mainHeaderViewFeature,
		mainViewFeature,
		remoteViewFeature,
		remoteBranchViewFeature,
		branchViewFeature,
		batchBranchViewFeature,
		suggestBranchViewFeature,
		stashViewFeature,
		commitViewFeature,
		dynamicViewFeature,
		keybindingsViewFeature,
	}
)

// set the layout and create views with their default size, name etc. values
// TODO: window sizes can be handled better
func (gui *Gui) focusLayout(g *gocui.Gui) error {
	return gui.overviewLayout(g)
}

// evolve the layout to focus layout and focus to commitview also initialize
// some stuff
func (gui *Gui) focusToRepository(g *gocui.Gui, v *gocui.View) error {
	r := gui.getSelectedRepository()
	if r == nil {
		return nil
	}
	if err := gui.renderRepositoryDetails(r); err != nil {
		return err
	}
	targetViewName := commitViewFeature.Name
	if v != nil {
		switch v.Name() {
		case mainViewFeature.Name, commitViewFeature.Name, dynamicViewFeature.Name, remoteViewFeature.Name, branchViewFeature.Name, stashViewFeature.Name:
			targetViewName = v.Name()
		}
	}
	return gui.focusToView(targetViewName)
}

// return back to overview layout
func (gui *Gui) focusBackToMain(g *gocui.Gui, v *gocui.View) error {
	return gui.focusToView(mainViewFeature.Name)
}

// focus to next view
func (gui *Gui) nextFocusView(g *gocui.Gui, v *gocui.View) error {
	return gui.nextViewOfGroup(g, v, paneViews)
}

// focus to previous view
func (gui *Gui) previousFocusView(g *gocui.Gui, v *gocui.View) error {
	return gui.previousViewOfGroup(g, v, paneViews)
}

func (gui *Gui) nextPaneView(g *gocui.Gui, v *gocui.View) error {
	return gui.nextViewOfGroup(g, v, paneViews)
}

func (gui *Gui) previousPaneView(g *gocui.Gui, v *gocui.View) error {
	return gui.previousViewOfGroup(g, v, paneViews)
}

func (gui *Gui) focusMainPane(g *gocui.Gui, v *gocui.View) error {
	return gui.focusToView(mainViewFeature.Name)
}

func (gui *Gui) focusBranchPane(g *gocui.Gui, v *gocui.View) error {
	return gui.focusToView(branchViewFeature.Name)
}
