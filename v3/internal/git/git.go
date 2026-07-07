package git

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"strings"
)

// ErrNotInstalled is returned when git is not found in PATH.
var ErrNotInstalled = errors.New("git is not installed; please install git from https://git-scm.com")

func isNotFound(err error) bool {
	var execErr *exec.Error
	return errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound)
}

// redactArgs returns a copy of args with any URL credentials (user:pass@host)
// replaced by user:***@host so tokens are not leaked in error messages.
func redactArgs(args []string) []string {
	out := make([]string, len(args))
	for i, a := range args {
		if u, err := url.Parse(a); err == nil && u.User != nil {
			u.User = url.UserPassword(u.User.Username(), "***")
			a = u.String()
		}
		out[i] = a
	}
	return out
}

func run(args ...string) error {
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		if isNotFound(err) {
			return ErrNotInstalled
		}
		return fmt.Errorf("git %s: %w\n%s", strings.Join(redactArgs(args), " "), err, bytes.TrimSpace(out))
	}
	return nil
}

func output(args ...string) (string, error) {
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		if isNotFound(err) {
			return "", ErrNotInstalled
		}
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(redactArgs(args), " "), err, bytes.TrimSpace(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// HeadHash returns the short (8-character) commit hash of HEAD in dir.
func HeadHash(dir string) (string, error) {
	hash, err := output("-C", dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	if len(hash) < 8 {
		return "", fmt.Errorf("git rev-parse returned unexpected output %q", hash)
	}
	return hash[:8], nil
}

// Clone clones url into dir. If tag is non-empty, checks out that tag or branch.
func Clone(url, dir, tag string) error {
	args := []string{"clone", "--quiet"}
	if tag != "" {
		args = append(args, "--branch", tag)
	}
	args = append(args, url, dir)
	return run(args...)
}

// stream runs git with args, sending live stdout+stderr to progress (typically
// os.Stderr) instead of buffering, so a long clone shows "Receiving objects..."
// as it goes rather than looking frozen.
func stream(progress io.Writer, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = progress
	cmd.Stderr = progress // git writes clone/checkout progress to stderr
	if err := cmd.Run(); err != nil {
		if isNotFound(err) {
			return ErrNotInstalled
		}
		return fmt.Errorf("git %s: %w", strings.Join(redactArgs(args), " "), err)
	}
	return nil
}

// CloneShallow clones url into dir at depth 1, streaming git's progress to
// progress (typically os.Stderr) so a large clone never looks frozen. Unlike
// Clone it does not buffer output, so the user sees "Receiving objects..." live.
// If tag is non-empty, that branch or tag is checked out.
func CloneShallow(url, dir, tag string, progress io.Writer) error {
	args := []string{"clone", "--depth", "1", "--progress"}
	if tag != "" {
		args = append(args, "--branch", tag)
	}
	args = append(args, url, dir)
	return stream(progress, args...)
}

// CloneSparse does a shallow, partial (blob:none) clone of url into dir and
// checks out only the given paths via sparse-checkout. For a docs subtree of a
// large monorepo this fetches just those paths' blobs, a fraction of a full
// shallow clone. It needs a server that supports partial clone (GitHub does) and
// git 2.25+; callers should fall back to CloneShallow if it errors. If tag is
// non-empty, that branch or tag is checked out.
func CloneSparse(url, dir, tag string, paths []string, progress io.Writer) error {
	args := []string{"clone", "--depth", "1", "--filter=blob:none", "--sparse", "--progress"}
	if tag != "" {
		args = append(args, "--branch", tag)
	}
	args = append(args, url, dir)
	if err := stream(progress, args...); err != nil {
		return err
	}
	setArgs := append([]string{"-C", dir, "sparse-checkout", "set"}, paths...)
	return stream(progress, setArgs...)
}

// Init initializes a new git repository at dir.
func Init(dir string) error {
	return run("-C", dir, "init", "--quiet")
}

// RemoteAdd adds a named remote to the repository at dir.
func RemoteAdd(dir, name, url string) error {
	return run("-C", dir, "remote", "add", name, url)
}

// AddAll stages all files in the repository at dir.
func AddAll(dir string) error {
	return run("-C", dir, "add", ".")
}
