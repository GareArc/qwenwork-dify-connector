package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSavedLoginReadsOnlyTopLevelServerAndEmail(t *testing.T) {
	dir := t.TempDir()
	content := "server: 'https://dify.example.invalid'\r\nemail: \"user@example.invalid\"\naccount:\n  email: nested@example.invalid\ntoken_storage: file\n"
	if err := os.WriteFile(filepath.Join(dir, loginFileName), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	login := readSavedLogin(dir)
	if login.Server != "https://dify.example.invalid" || login.Email != "user@example.invalid" {
		t.Fatalf("login = %+v", login)
	}
	if got := login.summary(); got != "user@example.invalid · https://dify.example.invalid" {
		t.Errorf("summary = %q", got)
	}
}

func TestReadSavedLoginWithoutAFileIsEmpty(t *testing.T) {
	if login := readSavedLogin(t.TempDir()); login != (savedLogin{}) || login.summary() != "" {
		t.Errorf("login = %+v", login)
	}
}

func TestValidateConfigDirRejectsRelativeAndControlCharacters(t *testing.T) {
	absolute, err := filepath.Abs("difyctl")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := validateConfigDir(absolute + string(filepath.Separator)); err != nil || got != absolute {
		t.Errorf("validateConfigDir(%q) = %q, %v", absolute, got, err)
	}
	for _, bad := range []string{"", "relative/difyctl", absolute + "\n"} {
		if _, err := validateConfigDir(bad); err == nil {
			t.Errorf("validateConfigDir(%q) accepted", bad)
		}
	}
}

func TestLoadSettingsDropsInvalidSavedValues(t *testing.T) {
	profileRoot = t.TempDir()
	t.Cleanup(func() { profileRoot = "" })
	raw := `{"server":"http://insecure.invalid","shared_config_dir":"relative/dir","future_field":true}`
	if err := os.WriteFile(filepath.Join(profileRoot, "settings.json"), []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	if got := loadSettings(); got != (connectorSettings{}) {
		t.Errorf("settings = %+v, want empty", got)
	}
	dir, err := difyctlConfigDir()
	if err != nil || dir != filepath.Join(profileRoot, "difyctl") {
		t.Errorf("config dir = %q, %v; want the private profile", dir, err)
	}
}
