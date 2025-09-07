package git

import (
	"os"
	"os/exec"
	"path/filepath"
)

func InRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func CommitWithMessage(msg string) error {
	tmp, err := os.CreateTemp("", "gommit-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(msg); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	cmd := exec.Command("git", "commit", "-F", tmp.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func WriteCommitEditMsg(path, msg string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(msg), 0o644)
}
