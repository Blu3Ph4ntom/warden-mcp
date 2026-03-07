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
	target := requestedPath
	if !filepath.IsAbs(target) {
		target = filepath.Join(rootAbs, target)
	}
	targetAbs, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return "", fmt.Errorf("resolve target path: %w", err)
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
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
	return targetAbs, nil
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
