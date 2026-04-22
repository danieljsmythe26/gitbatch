package command

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/isacikgoz/gitbatch/internal/git"
	"github.com/stretchr/testify/require"
)

var (
	testPushopts1 = &PushOptions{
		RemoteName: "origin",
	}

	testPushopts2 = &PushOptions{
		RemoteName: "origin",
		Force:      true,
	}
)

func TestPushWithGit(t *testing.T) {
	th := git.InitTestRepositoryFromLocal(t)
	defer th.CleanUp(t)
	configureLocalPushRemote(t, th.RepoPath)

	var tests = []struct {
		inp1 *git.Repository
		inp2 *PushOptions
	}{
		{th.Repository, testPushopts1},
		{th.Repository, testPushopts2},
	}
	for _, test := range tests {
		err := pushWithGit(test.inp1, test.inp2)
		require.NoError(t, err)
	}
}

func configureLocalPushRemote(t *testing.T, repoPath string) string {
	t.Helper()

	remotePath := filepath.Join(filepath.Dir(repoPath), "origin.git")
	err := os.MkdirAll(remotePath, 0755)
	require.NoError(t, err)

	cmd := exec.Command("git", "init", "--bare", remotePath)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	cmd = exec.Command("git", "remote", "set-url", "origin", remotePath)
	cmd.Dir = repoPath
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	return remotePath
}
