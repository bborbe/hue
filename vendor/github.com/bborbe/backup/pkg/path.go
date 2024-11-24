package pkg

import (
	"context"
	"os"
	"path"

	"github.com/bborbe/errors"
)

type Paths []Path

type Path string

func (f Path) String() string {
	return string(f)
}

func (f Path) Join(elem ...string) Path {
	return Path(path.Join(append([]string{f.String()}, elem...)...))
}

func (f Path) Exists(ctx context.Context) (bool, error) {
	if _, err := os.Stat(f.String()); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f Path) Remove(ctx context.Context) error {
	if err := os.Remove(f.String()); err != nil {
		return errors.Wrapf(ctx, err, "remove failed")
	}
	return nil
}

func (f Path) Rename(ctx context.Context, path Path) error {
	if err := os.Rename(f.String(), path.String()); err != nil {
		return errors.Wrapf(ctx, err, "rename failed")
	}
	return nil
}

func (f Path) List(ctx context.Context) (Paths, error) {
	dirEntries, err := os.ReadDir(f.String())
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "list failed")
	}
	var result Paths
	for _, dir := range dirEntries {
		result = append(result, f.Join(dir.Name()))
	}
	return result, nil
}
