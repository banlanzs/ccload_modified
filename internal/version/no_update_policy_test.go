package version

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNoUpdateCheckerPolicy(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))

	for _, tc := range []struct {
		relPath    string
		forbidden  string
		errorLabel string
	}{
		{relPath: "main.go", forbidden: "version.StartChecker()", errorLabel: "startup update checker call"},
		{relPath: filepath.Join("internal", "app", "admin_stats.go"), forbidden: "version.GetUpdateInfo()", errorLabel: "public version update metadata"},
		{relPath: filepath.Join("web", "assets", "js", "ui.js"), forbidden: "versionInfo.has_update", errorLabel: "frontend update indicator"},
		{relPath: filepath.Join("internal", "version", "checker.go"), forbidden: "api.github.com/repos/caidaoli/ccLoad/releases/latest", errorLabel: "github releases update probe"},
	} {
		fullPath := filepath.Join(repoRoot, tc.relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("read %s failed: %v", tc.relPath, err)
		}
		if strings.Contains(string(content), tc.forbidden) {
			t.Fatalf("forbidden %s found in %s", tc.errorLabel, tc.relPath)
		}
	}
}
