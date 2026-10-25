package utils

import (
	"os"
	"strings"
)

// rpm-ostree and other image builders keep package-created accounts in /usr/lib,
// so reading /etc alone misses them and groupadd then fails with "already exists".
var (
	groupFiles  = []string{"/etc/group", "/usr/lib/group"}
	passwdFiles = []string{"/etc/passwd", "/usr/lib/passwd"}
)

func HasGroup(groupName string) bool {
	return HasGroupIn(groupName, groupFiles...)
}

func HasGroupIn(groupName string, paths ...string) bool {
	return HasGroupData(groupName, ReadAccountDB(paths...))
}

func GroupDB() string {
	return ReadAccountDB(groupFiles...)
}

func PasswdDB() string {
	return ReadAccountDB(passwdFiles...)
}

func ReadAccountDB(paths ...string) string {
	var b strings.Builder
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		b.Write(data)
		if len(data) > 0 && data[len(data)-1] != '\n' {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func HasGroupData(groupName, data string) bool {
	prefix := groupName + ":"
	for line := range strings.SplitSeq(data, "\n") {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func FindGroupData(data string, candidates ...string) (string, bool) {
	for _, candidate := range candidates {
		if HasGroupData(candidate, data) {
			return candidate, true
		}
	}
	return "", false
}
