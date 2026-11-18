//go:build darwin || linux

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func defaultProfileRoot() (string, error) {
	home, err := os.UserHomeDir()
	return filepath.Join(home, ".config", "qwenwork-dify-connector"), err
}

func lockFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func lockBusy(err error) bool {
	return errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)
}

func unlockFile(file *os.File) {
	_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

func openBrowser(address string) error {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", address).Run()
	}
	// Some Linux desktop openers stay attached for the browser's lifetime.
	// Reap the opener without blocking portal completion or cancellation.
	command := exec.Command("xdg-open", address)
	if err := command.Start(); err != nil {
		return err
	}
	finished := make(chan error, 1)
	go func() { finished <- command.Wait() }()
	select {
	case err := <-finished:
		return err
	case <-time.After(250 * time.Millisecond):
		return nil
	}
}

func executableName(name string) string { return name }

const shellEnvMarker = "__DIFY_QWEN_ENV__"

// Reads variables as the user's terminal sees them. Interactive login shells can
// print banners, so the values follow a marker, one per line.
func loginShellEnv(names ...string) map[string]string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return nil
	}
	script := "printf '" + shellEnvMarker + "\\n'"
	for _, name := range names {
		script += "; printf '%s\\n' \"$" + name + "\""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, shell, "-ilc", script)
	command.WaitDelay = time.Second
	out, err := command.Output()
	if err != nil {
		return nil
	}
	at := strings.LastIndex(string(out), shellEnvMarker+"\n")
	if at < 0 {
		return nil
	}
	values := strings.Split(string(out[at+len(shellEnvMarker)+1:]), "\n")
	env := map[string]string{}
	for i, name := range names {
		if i < len(values) {
			env[name] = values[i]
		}
	}
	return env
}

func defaultCLIConfigDir(shellXDG string) string {
	if runtime.GOOS == "linux" {
		xdg := shellXDG
		if xdg == "" {
			xdg = os.Getenv(envXDGConfigHome)
		}
		if xdg != "" {
			return filepath.Join(xdg, "difyctl")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "difyctl")
}
