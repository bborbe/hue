package v1

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func ParseBackupExcludesFromString(value string) BackupExcludes {
	return ParseBackupExcludes(strings.FieldsFunc(value, func(r rune) bool {
		return r == ','
	}))
}

func ParseBackupExcludes(values []string) BackupExcludes {
	result := make(BackupExcludes, len(values))
	for i, value := range values {
		result[i] = BackupExclude(value)
	}
	return result
}

type BackupExcludes []BackupExclude

func (a BackupExcludes) Len() int { return len(a) }
func (a BackupExcludes) Less(i, j int) bool {
	return strings.Compare(strings.ToLower(a[i].String()), strings.ToLower(a[j].String())) < 0
}
func (a BackupExcludes) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a BackupExcludes) Strings() []string {
	result := make([]string, len(a))
	for i, aa := range a {
		result[i] = aa.String()
	}
	return result
}

func (a BackupExcludes) Sorted() BackupExcludes {
	sort.Sort(a)
	return a
}

func (a BackupExcludes) Bytes() []byte {
	buf := &bytes.Buffer{}
	for _, aa := range a {
		fmt.Fprintln(buf, aa)
	}
	return buf.Bytes()
}

type BackupExclude string

func (f BackupExclude) String() string {
	return string(f)
}
