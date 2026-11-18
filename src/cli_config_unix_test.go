//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoginShellEnvSkipsBannersAndKeepsEmptyValues(t *testing.T) {
	shell := filepath.Join(t.TempDir(), "shell")
	script := "#!/bin/sh\necho 'Welcome banner __DIFY_QWEN'\nexec /bin/sh -c \"$2\"\n"
	if err := os.WriteFile(shell, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHELL", shell)
	t.Setenv(envDifyConfigDir, "/Users/someone/dify config")
	t.Setenv(envXDGConfigHome, "")
	env := loginShellEnv(envDifyConfigDir, envXDGConfigHome)
	if env[envDifyConfigDir] != "/Users/someone/dify config" {
		t.Errorf("DIFY_CONFIG_DIR = %q", env[envDifyConfigDir])
	}
	if value, ok := env[envXDGConfigHome]; !ok || value != "" {
		t.Errorf("XDG_CONFIG_HOME = %q, %v; want an empty value", value, ok)
	}
}

func TestLoginShellEnvGivesNothingWhenTheShellFails(t *testing.T) {
	t.Setenv("SHELL", filepath.Join(t.TempDir(), "missing-shell"))
	if env := loginShellEnv(envDifyConfigDir); len(env) != 0 {
		t.Errorf("env = %v, want none", env)
	}
}
