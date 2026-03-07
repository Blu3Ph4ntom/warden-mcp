package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveWorkspacePathAllowsMarkdownInsideWorkspace(t *testing.T) {
	root := t.TempDir()
	target, err := ResolveWorkspacePath(root, filepath.Join(".agent", "PLAN.md"), ".md")
	if err != nil {
		t.Fatalf("expected workspace-relative markdown path to pass, got %v", err)
	}
	if !strings.HasSuffix(target, filepath.Join(".agent", "PLAN.md")) {
		t.Fatalf("unexpected resolved path %s", target)
	}
}

func TestResolveWorkspacePathRejectsTraversalAndWrongExtension(t *testing.T) {
	root := t.TempDir()
	if _, err := ResolveWorkspacePath(root, filepath.Join("..", "outside.md"), ".md"); err == nil {
		t.Fatal("expected traversal path to fail")
	}
	if _, err := ResolveWorkspacePath(root, "plan.txt", ".md"); err == nil {
		t.Fatal("expected wrong extension to fail")
	}
}

func TestResolveWorkspacePathRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.md"), []byte("secret"), 0o644); err != nil {
		t.Fatalf("write outside file failed: %v", err)
	}
	linkPath := filepath.Join(root, "linked")
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if _, err := ResolveWorkspacePath(root, filepath.Join("linked", "secret.md"), ".md"); err == nil {
		t.Fatal("expected symlink escape to fail")
	}
}

func TestFingerprintFileAndRedactionHelpers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "PLAN.md")
	if err := os.WriteFile(path, []byte("token sk-1234567890abcdef\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	fingerprint, err := FingerprintFile(path, 1024)
	if err != nil {
		t.Fatalf("fingerprint failed: %v", err)
	}
	if fingerprint.Size == 0 || fingerprint.SHA256 == "" {
		t.Fatalf("unexpected fingerprint %+v", fingerprint)
	}
	redacted := RedactSecretLikeText("use sk-1234567890abcdef now")
	if strings.Contains(redacted, "1234567890abcdef") || !strings.Contains(redacted, "***redacted***") {
		t.Fatalf("expected secret-like text to be redacted, got %s", redacted)
	}
}
