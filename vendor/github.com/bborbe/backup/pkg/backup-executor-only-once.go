package pkg

import (
	"context"
	stderrors "errors"
	"sync"

	v1 "github.com/bborbe/backup/k8s/apis/backup.benjamin-borbe.de/v1"
)

var AlreadyRunningError = stderrors.New("backup already running")

func NewBackupExectuorOnlyOnce(
	backupExectuor BackupExectuor,
) BackupExectuor {
	return &backupExectuorOnlyOnce{
		backupExectuor: backupExectuor,
	}
}

type backupExectuorOnlyOnce struct {
	mux            sync.Mutex
	running        bool
	backupExectuor BackupExectuor
}

func (b *backupExectuorOnlyOnce) Backup(ctx context.Context, target v1.BackupSpec) error {
	b.mux.Lock()
	if b.running {
		b.mux.Unlock()
		return AlreadyRunningError
	}
	b.running = true
	b.mux.Unlock()
	err := b.backupExectuor.Backup(ctx, target)
	b.mux.Lock()
	b.running = false
	b.mux.Unlock()
	return err
}
