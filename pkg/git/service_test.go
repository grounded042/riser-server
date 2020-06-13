package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewRepo_ValidatesSshKeyPath(t *testing.T) {
	repoSettings := RepoSettings{
		SSHKeyPath: "/bogus/path",
	}
	repo, err := NewRepo(repoSettings)

	assert.Nil(t, repo)
	assert.Equal(t, "Error reading ssh key: stat /bogus/path: no such file or directory", err.Error())
}

func Test_NewRepo_ValidatesUrlIsSsh_WhenSshKeyPathIsSet(t *testing.T) {
	repoSettings := RepoSettings{
		URL:        "https://not-ssh.org",
		SSHKeyPath: "/bogus/path",
	}
	repo, err := NewRepo(repoSettings)

	assert.Nil(t, repo)
	assert.Equal(t, "Cannot use both an https git url and specify an SSH key. Either use an SSH url or remove the key", err.Error())
}

func Test_buildGitCmd(t *testing.T) {
	repoSettings := &RepoSettings{
		LocalGitDir: "/tmp/git",
	}
	r := repo{settings: repoSettings}
	ctx := context.Background()
	cmd := r.buildGitCmd(ctx, "status")

	assert.Equal(t, []string{"git", "status"}, cmd.Args)
	assert.Equal(t, "/tmp/git", cmd.Dir)
}
