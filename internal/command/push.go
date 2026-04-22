package command

import (
	"strings"

	gerr "github.com/isacikgoz/gitbatch/internal/errors"
	"github.com/isacikgoz/gitbatch/internal/git"
)

// PushOptions defines the rules for push operation.
type PushOptions struct {
	// Name of the remote to push to. Defaults to origin.
	RemoteName string
	// ReferenceName remote branch to push to. If empty, pushes the current branch.
	ReferenceName string
	// Force allows the push to update the remote branch even when it is not a fast-forward.
	Force bool
	// SetUpstream adds upstream tracking while pushing.
	SetUpstream bool
	// Mode is the command mode.
	CommandMode Mode
}

// Push updates remote refs using local refs.
func Push(r *git.Repository, o *PushOptions) error {
	switch o.CommandMode {
	case ModeLegacy, ModeNative:
		return pushWithGit(r, o)
	default:
		return pushWithGit(r, o)
	}
}

func pushWithGit(r *git.Repository, options *PushOptions) error {
	args := []string{"push"}
	if options.Force {
		args = append(args, "-f")
	}
	if options.SetUpstream {
		args = append(args, "-u")
	}
	if len(options.RemoteName) > 0 {
		args = append(args, options.RemoteName)
	}
	if len(options.ReferenceName) > 0 {
		args = append(args, options.ReferenceName)
	}

	out, err := Run(r.AbsPath, "git", args)
	if err != nil {
		return gerr.ParseGitError(out, err)
	}

	r.SetWorkStatus(git.Success)
	switch {
	case len(options.RemoteName) > 0 && len(options.ReferenceName) > 0:
		r.State.Message = "pushed " + options.ReferenceName + " to " + options.RemoteName
	case len(options.RemoteName) > 0:
		r.State.Message = "pushed to " + options.RemoteName
	default:
		r.State.Message = "push complete"
	}
	if strings.Contains(out, "Everything up-to-date") {
		r.State.Message = "already up-to-date"
	}
	return r.Refresh()
}
