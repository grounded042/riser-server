package git

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/riser-platform/riser-server/pkg/core"
	"github.com/riser-platform/riser-server/pkg/util"
)

var (
	// ErrNoChanges indicates that there were no changes to commit
	ErrNoChanges = errors.New("no changes to commit")
)

const (
	// TODO: After adding auth consider making this dynamic e.g. "riser-server (initiated by johndoe@acme.org)"
	commitName  = "riser-server"
	commitEmail = "riser-server@tempuri.org"
	remoteName  = "origin"

	// TODO: Consider making configurable - the main scenario is large repos that take a long time for the initial clone
	gitExecTimeoutSeconds = 30 * time.Second
)

type RepoSettings struct {
	URL         string
	SSHKeyPath  string
	Branch      string
	LocalGitDir string
}

type Repo interface {
	Commit(message string, files []core.ResourceFile) error
	Push() error
	ResetHardRemote() error
	// Lock locks the repo. Be sure to call Unlock when your work is completed.
	Lock()
	// Unlock unlocks the repo.
	Unlock()
}

type repo struct {
	settings *RepoSettings
	sync.Mutex
}

// NewRepo creates a new reference to a repo. There should only be one instance running per git folder.
// WARNING: any pending changes or local commits will be lost
func NewRepo(repoSettings RepoSettings) (Repo, error) {
	repo := &repo{
		settings: &repoSettings,
		Mutex:    sync.Mutex{},
	}

	if repoSettings.SSHKeyPath != "" {
		if strings.HasPrefix(repoSettings.URL, "https://") {
			return nil, errors.New("Cannot use both an https git url and specify an SSH key. Either use an SSH url or remove the key")
		}

		_, err := os.Stat(repoSettings.SSHKeyPath)
		if err != nil {
			return nil, errors.Wrap(err, "Error reading ssh key")
		}
	}

	err := repo.init()
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *repo) Commit(message string, files []core.ResourceFile) error {
	err := processFiles(repo.settings.LocalGitDir, files)
	if err != nil {
		return err
	}
	err = repo.addAll()
	if err != nil {
		return err
	}
	err = repo.execGitCmd("commit", "-m", message, "--author", fmt.Sprintf("%s <%s>", commitName, commitEmail))
	if err != nil && isNoChangesErr(err) {
		return ErrNoChanges
	}
	return err
}

func (repo *repo) addAll() error {
	return repo.execGitCmd("add", "--all")
}

func (repo *repo) Push() error {
	return repo.execGitCmd("push")
}

func (repo *repo) init() error {
	err := util.EnsureDir(util.EnsureTrailingSlash(repo.settings.LocalGitDir), workspaceFilePerm)
	if err != nil {
		return errors.Wrap(err, "error ensuring git dir")
	}
	files, err := ioutil.ReadDir(repo.settings.LocalGitDir)
	if err != nil {
		return errors.Wrap(err, "error reading git dir")
	}
	if len(files) == 0 {
		return repo.clone()
	}

	err = repo.fetch()
	if err != nil {
		return err
	}

	// Remote is always the source of truth. If a previous process aborted before pushing a commit, it is considered a failed transaction
	err = repo.clean()
	if err != nil {
		return err
	}

	return repo.ResetHardRemote()
}

func (repo *repo) clone() error {
	return repo.execGitCmd("clone", "--branch", repo.settings.Branch, "--single-branch", "--depth=1", repo.settings.URL, repo.settings.LocalGitDir)
}

func (repo *repo) fetch() error {
	return repo.execGitCmd("fetch", "-f", remoteName, repo.settings.Branch)
}

func (repo *repo) clean() error {

	return repo.execGitCmd("clean", "-xdf")
}

// ResetHardRemote ensures that the remote is up-to-date. Pending commits will be lost.
func (repo *repo) ResetHardRemote() error {
	// Always fetch before resetting to the remote to ensure that we're up-to-date
	err := repo.fetch()
	if err != nil {
		return err
	}

	return repo.execGitCmd("reset", "--hard", fmt.Sprintf("%s/%s", remoteName, repo.settings.Branch))
}

func (repo *repo) execGitCmd(args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), gitExecTimeoutSeconds)
	defer cancel()
	cmd := repo.buildGitCmd(ctx, args...)
	_, err := execWithContext(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("git %s", args))
	}

	return nil
}

func (repo *repo) buildGitCmd(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repo.settings.LocalGitDir
	return cmd
}
