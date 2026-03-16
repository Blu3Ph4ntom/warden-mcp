package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const WorkspaceRootOverrideEnv = "WARDEN_WORKSPACE_ROOT"
const AllowUnsafeWorkspaceFallbackEnv = "WARDEN_ALLOW_UNSAFE_WORKSPACE_FALLBACK"

type WorkspaceRootResolution struct {
	Root   string
	Source string
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)sk-[a-z0-9]{16,}`),
	regexp.MustCompile(`ghp_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
}

type FileFingerprint struct {
	Path   string
	Size   int64
	SHA256 string
}

func ResolveProcessWorkspaceRoot() (WorkspaceRootResolution, error) {
	return resolveProcessWorkspaceRoot(os.Getenv, os.Getwd, os.UserHomeDir)
}

func ResolveWorkspaceRoot(root string) (string, error) {
	return normalizeWorkspaceRoot(root)
}

func IsUnsafeWorkspaceRootPath(root string) bool {
	resolved, err := normalizeWorkspaceRoot(root)
	if err != nil {
		return true
	}
	return isUnsafeWorkspaceRoot(resolved, os.Getenv)
}

func ResolveWorkspacePath(workspaceRoot, requestedPath string, allowedExts ...string) (string, error) {
	if strings.TrimSpace(workspaceRoot) == "" {
		return "", fmt.Errorf("workspace root is required")
	}
	if strings.TrimSpace(requestedPath) == "" {
		return "", fmt.Errorf("path is required")
	}
	rootAbs, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root: %w", err)
	}
	rootResolved, err := evalSymlinksAllowMissing(rootAbs)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root symlinks: %w", err)
	}
	target := requestedPath
	if !filepath.IsAbs(target) {
		target = filepath.Join(rootAbs, target)
	}
	targetAbs, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return "", fmt.Errorf("resolve target path: %w", err)
	}
	targetResolved, err := evalSymlinksAllowMissing(targetAbs)
	if err != nil {
		return "", fmt.Errorf("resolve target symlinks: %w", err)
	}
	rel, err := filepath.Rel(rootResolved, targetResolved)
	if err != nil {
		return "", fmt.Errorf("compare target path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes workspace: %s", requestedPath)
	}
	if len(allowedExts) > 0 {
		ext := strings.ToLower(filepath.Ext(targetAbs))
		allowed := false
		for _, candidate := range allowedExts {
			if strings.ToLower(candidate) == ext {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("path must use one of the allowed extensions: %v", allowedExts)
		}
	}
	return targetResolved, nil
}

func resolveProcessWorkspaceRoot(getenv func(string) string, getwd func() (string, error), userHomeDir func() (string, error)) (WorkspaceRootResolution, error) {
	if override := strings.TrimSpace(getenv(WorkspaceRootOverrideEnv)); override != "" {
		resolved, err := normalizeWorkspaceRoot(override)
		if err != nil {
			return WorkspaceRootResolution{}, fmt.Errorf("resolve %s: %w", WorkspaceRootOverrideEnv, err)
		}
		return WorkspaceRootResolution{Root: resolved, Source: "env"}, nil
	}
	cwd, err := getwd()
	if err != nil {
		return WorkspaceRootResolution{}, err
	}
	resolvedCWD, err := normalizeWorkspaceRoot(cwd)
	if err != nil {
		return WorkspaceRootResolution{}, fmt.Errorf("resolve working directory: %w", err)
	}
	if !isUnsafeWorkspaceRoot(resolvedCWD, getenv) {
		return WorkspaceRootResolution{Root: resolvedCWD, Source: "cwd"}, nil
	}
	if strings.TrimSpace(getenv(AllowUnsafeWorkspaceFallbackEnv)) != "1" {
		return WorkspaceRootResolution{}, fmt.Errorf("unsafe workspace root %q; set %s to your repo root or set %s=1 to opt into the legacy shared fallback", resolvedCWD, WorkspaceRootOverrideEnv, AllowUnsafeWorkspaceFallbackEnv)
	}
	home, err := userHomeDir()
	if err != nil {
		return WorkspaceRootResolution{}, fmt.Errorf("resolve fallback home directory: %w", err)
	}
	resolvedHome, err := normalizeWorkspaceRoot(filepath.Join(home, ".warden-mcp", "workspaces", "default"))
	if err != nil {
		return WorkspaceRootResolution{}, fmt.Errorf("resolve fallback workspace root: %w", err)
	}
	return WorkspaceRootResolution{Root: resolvedHome, Source: "home_fallback"}, nil
}

func normalizeWorkspaceRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("workspace root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return evalSymlinksAllowMissing(abs)
}

func isUnsafeWorkspaceRoot(root string, getenv func(string) string) bool {
	cleaned := filepath.Clean(root)
	if cleaned == filepath.Dir(cleaned) {
		return true
	}
	if windir := strings.TrimSpace(getenv("WINDIR")); windir != "" {
		resolvedWindir, err := normalizeWorkspaceRoot(windir)
		if err == nil && pathWithinOrEqual(resolvedWindir, cleaned) {
			return true
		}
	}
	for _, envName := range []string{"ProgramFiles", "ProgramFiles(x86)", "ProgramData"} {
		candidate := strings.TrimSpace(getenv(envName))
		if candidate == "" {
			continue
		}
		resolvedCandidate, err := normalizeWorkspaceRoot(candidate)
		if err == nil && pathWithinOrEqual(resolvedCandidate, cleaned) {
			return true
		}
	}
	return false
}

func pathWithinOrEqual(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || rel == "" || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}

func evalSymlinksAllowMissing(path string) (string, error) {
	cleaned := filepath.Clean(path)
	missing := make([]string, 0)
	probe := cleaned
	for {
		if _, err := os.Lstat(probe); err == nil {
			resolved, err := filepath.EvalSymlinks(probe)
			if err != nil {
				return "", err
			}
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return filepath.Clean(resolved), nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			return filepath.Abs(cleaned)
		}
		missing = append(missing, filepath.Base(probe))
		probe = parent
	}
}

func FingerprintFile(path string, maxBytes int64) (FileFingerprint, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileFingerprint{}, err
	}
	if maxBytes > 0 && info.Size() > maxBytes {
		return FileFingerprint{}, fmt.Errorf("file exceeds max size of %d bytes", maxBytes)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return FileFingerprint{}, err
	}
	hash := sha256.Sum256(content)
	return FileFingerprint{Path: path, Size: info.Size(), SHA256: hex.EncodeToString(hash[:])}, nil
}

func RedactSecretLikeText(input string) string {
	redacted := input
	for _, pattern := range secretPatterns {
		redacted = pattern.ReplaceAllStringFunc(redacted, func(value string) string {
			return preservePrefix(value, 4) + "***redacted***"
		})
	}
	return redacted
}

func preservePrefix(value string, keep int) string {
	if len(value) <= keep {
		return ""
	}
	return value[:keep]
}
