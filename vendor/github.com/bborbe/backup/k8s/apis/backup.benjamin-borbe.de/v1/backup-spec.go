package v1

import (
	"context"
	"strconv"

	"github.com/bborbe/collection"
	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
)

type BackupSpecs []BackupSpec

type BackupHost string

func (f BackupHost) String() string {
	return string(f)
}

type BackupPort int

func (f BackupPort) Int() int {
	return int(f)
}

func (f BackupPort) String() string {
	return strconv.Itoa(f.Int())
}

type BackupUser string

func (f BackupUser) String() string {
	return string(f)
}

// BackupSpec is the spec for a Foo resource
type BackupSpec struct {
	Host     BackupHost     `json:"host" yaml:"host"`
	Port     BackupPort     `json:"port" yaml:"port"`
	User     BackupUser     `json:"user" yaml:"user"`
	Dirs     BackupDirs     `json:"dirs" yaml:"dirs"`
	Excludes BackupExcludes `json:"excludes" yaml:"excludes"`
}

func (a BackupSpec) Equal(backup BackupSpec) bool {
	if a.Host != backup.Host {
		return false
	}
	if a.Port != backup.Port {
		return false
	}
	if a.User != backup.User {
		return false
	}
	if collection.Equal(a.Dirs.Sorted(), backup.Dirs.Sorted()) == false {
		return false
	}
	if collection.Equal(a.Excludes.Sorted(), backup.Excludes.Sorted()) == false {
		return false
	}
	return true
}

func (a BackupSpec) Validate(ctx context.Context) error {
	if a.Host == "" {
		return errors.Wrap(ctx, validation.Error, "Host is empty")
	}
	if a.Port <= 0 {
		return errors.Wrap(ctx, validation.Error, "Port is invalid")
	}
	if a.User == "" {
		return errors.Wrap(ctx, validation.Error, "User is empty")
	}
	return nil
}
