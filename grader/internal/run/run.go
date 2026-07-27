package run

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Manifest struct {
	ID       string   `json:"id"`
	Created  string   `json:"created"`
	Source   string   `json:"source"`
	Revision string   `json:"revision"`
	Seed     int64    `json:"seed"`
	Dir      string   `json:"dir"`
	Bugs     []string `json:"bugs"`
}

func Export(src, revision, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	archive := exec.Command("git", "-C", src, "archive", revision)
	extract := exec.Command("tar", "-x", "-C", dir)
	pipe, err := archive.StdoutPipe()
	if err != nil {
		return err
	}
	extract.Stdin = pipe
	archive.Stderr = os.Stderr
	extract.Stderr = os.Stderr
	if err := extract.Start(); err != nil {
		return err
	}
	if err := archive.Run(); err != nil {
		return err
	}
	return extract.Wait()
}

func Init(dir string) error {
	for _, args := range [][]string{
		{"init", "-q"},
		{"add", "-A"},
		{"-c", "user.name=solve-this", "-c", "user.email=solve-this@localhost", "commit", "-q", "-m", "initial import"},
	} {
		if _, err := Git(dir, args...); err != nil {
			return err
		}
	}
	return nil
}

func Command(dir string, env []string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, out.String())
	}
	return out.String(), nil
}

func Git(dir string, args ...string) (string, error) {
	out, err := Command(dir, nil, "git", args...)
	return strings.TrimSpace(out), err
}
