package gui

import (
	"github.com/isacikgoz/gitbatch/internal/command"
	"github.com/jroimartin/gocui"
)

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
		return nil
	}
	if err := command.Push(r, &command.PushOptions{
		RemoteName:  r.State.Remote.Name,
		CommandMode: command.ModeLegacy,
	}); err != nil {
		return err
	}
	return gui.initFocusStat(r)
}
