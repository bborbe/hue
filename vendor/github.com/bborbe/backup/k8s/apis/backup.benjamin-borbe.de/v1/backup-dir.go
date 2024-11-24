package v1

import (
	"sort"
	"strings"
)

func ParseBackupDirsFromString(value string) BackupDirs {
	return ParseBackupDirs(strings.FieldsFunc(value, func(r rune) bool {
		return r == ','
	}))
}

func ParseBackupDirs(values []string) BackupDirs {
	result := make(BackupDirs, len(values))
	for i, value := range values {
		result[i] = BackupDir(value)
	}
	return result
}

type BackupDirs []BackupDir

func (a BackupDirs) Len() int { return len(a) }
func (a BackupDirs) Less(i, j int) bool {
	return strings.Compare(strings.ToLower(a[i].String()), strings.ToLower(a[j].String())) < 0
}
func (a BackupDirs) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a BackupDirs) Strings() []string {
	result := make([]string, len(a))
	for i, aa := range a {
		result[i] = aa.String()
	}
	return result
}

func (a BackupDirs) Sorted() BackupDirs {
	sort.Sort(a)
	return a
}

type BackupDir string

func (f BackupDir) String() string {
	return string(f)
}
