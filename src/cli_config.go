package main

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	envDifyConfigDir = "DIFY_CONFIG_DIR"
	envXDGConfigHome = "XDG_CONFIG_HOME"
	loginFileName    = "login.yml"
)

// The user's own difyctl keeps its login in its config folder, wherever the binary
// is installed. QwenWork starts as a GUI app and does not see variables set in the
// user's shell, so the login shell is asked as well.
func detectCLIConfigDir() string {
	if dir := os.Getenv(envDifyConfigDir); dir != "" {
		return dir
	}
	shell := loginShellEnv(envDifyConfigDir, envXDGConfigHome)
	if dir := shell[envDifyConfigDir]; dir != "" {
		return dir
	}
	return defaultCLIConfigDir(shell[envXDGConfigHome])
}

type savedLogin struct {
	Server string
	Email  string
}

func (l savedLogin) summary() string {
	if l.Email == "" {
		return ""
	}
	return l.Email + " · " + l.Server
}

// Reads only the top-level server and email of difyctl's login.yml, for display and
// as a default server; the token is never in this file.
func readSavedLogin(dir string) savedLogin {
	raw, err := os.ReadFile(filepath.Join(dir, loginFileName))
	if err != nil {
		return savedLogin{}
	}
	var login savedLogin
	for _, line := range strings.Split(string(raw), "\n") {
		key, value, found := strings.Cut(strings.TrimRight(line, "\r"), ":")
		if !found {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch key {
		case "server":
			login.Server = value
		case "email":
			login.Email = value
		}
	}
	return login
}
