package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CommandExists(cmd string) bool {
	_, ok := LookCommand(cmd)
	return ok
}

// LookCommand also checks sbin, which openSUSE and Debian leave off an
// unprivileged user's PATH.
func LookCommand(cmd string) (string, bool) {
	if path, err := exec.LookPath(cmd); err == nil {
		return path, true
	}
	if strings.ContainsRune(cmd, os.PathSeparator) {
		return "", false
	}
	for _, dir := range append(userBinDirs(), systemSbinDirs...) {
		path := filepath.Join(dir, cmd)
		info, statErr := os.Stat(path)
		if statErr == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return path, true
		}
	}
	return "", false
}

var systemSbinDirs = []string{"/usr/local/sbin", "/usr/sbin", "/sbin"}

func EnvWithUserBinPath(env []string) []string {
	if env == nil {
		env = os.Environ()
	}

	out := append([]string(nil), env...)
	pathIndex := -1
	pathValue := ""
	for i, entry := range out {
		if strings.HasPrefix(entry, "PATH=") {
			pathIndex = i
			pathValue = strings.TrimPrefix(entry, "PATH=")
			break
		}
	}

	parts := filepath.SplitList(pathValue)
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		if part != "" {
			seen[part] = struct{}{}
		}
	}

	prepend := make([]string, 0, 2)
	for _, dir := range userBinDirs() {
		if dir == "" {
			continue
		}
		if _, ok := seen[dir]; ok {
			continue
		}
		prepend = append(prepend, dir)
		seen[dir] = struct{}{}
	}

	parts = append(prepend, parts...)
	newPath := "PATH=" + strings.Join(parts, string(os.PathListSeparator))
	if pathIndex >= 0 {
		out[pathIndex] = newPath
	} else {
		out = append(out, newPath)
	}
	return out
}

func userBinDirs() []string {
	dirs := []string{}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		dirs = append(dirs, filepath.Join(home, ".local", "bin"))
	}
	dirs = append(dirs, "/usr/local/bin")
	return dirs
}
