package gui

import (
	"github.com/isacikgoz/gitbatch/internal/command"
	"github.com/jroimartin/gocui"
)

func (gui *Gui) refreshSelectedRepositoryStatus(g *gocui.Gui, v *gocui.View) error {
	r := gui.getSelectedRepository()
	if r == nil || r.State == nil || r.State.Remote == nil {
		return nil
	}
	if err := command.Fetch(r, &command.FetchOptions{
		RemoteName:  r.State.Remote.Name,
		CommandMode: command.ModeNative,
	}); err != nil {
		return gui.openErrorView(g, "Refresh failed: "+err.Error(),
			"Close this dialog to return to repository status.",
			dynamicViewFeature.Name)
	}
	return gui.initFocusStat(r)
}

func (gui *Gui) pullSelectedRepository(g *gocui.Gui, v *gocui.View) error {
	r := gui.getSelectedRepository()
	if r == nil || r.State.Branch.Upstream == nil {
		return nil
	}
	if err := command.Pull(r, &command.PullOptions{
		RemoteName:  r.State.Remote.Name,
		CommandMode: command.ModeNative,
	}); err != nil {
		return err
	}
	return gui.initFocusStat(r)
}

func (gui *Gui) pushSelectedRepository(g *gocui.Gui, v *gocui.View) error {
	r := gui.getSelectedRepository()
	if r == nil || r.State.Branch.Upstream == nil {
		gui.setPushFeedback(r, false, "Push unavailable: branch is not tracking a remote branch.")
		return nil
	}
	if err := command.Push(r, &command.PushOptions{
		RemoteName:  r.State.Remote.Name,
		CommandMode: command.ModeLegacy,
	}); err != nil {
		gui.setPushFeedback(r, false, "Push failed: "+err.Error())
		return err
	}
	gui.setPushFeedback(r, true, r.State.Message)
	return gui.initFocusStat(r)
}
