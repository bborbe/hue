package pkg

import (
	"context"

	"github.com/bborbe/cron"
	libcron "github.com/bborbe/cron"
	"github.com/bborbe/errors"
	"github.com/bborbe/k8s"
	"github.com/bborbe/run"
	libsentry "github.com/bborbe/sentry"
	libtime "github.com/bborbe/time"
)

func CreateBackupCron(
	sentryClient libsentry.Client,
	backupExectuor BackupExectuor,
	kubeConfig string,
	namespace k8s.Namespace,
	cronExpression libcron.Expression,
) run.Func {
	return func(ctx context.Context) error {
		backupAction := CreateBackupAction(
			sentryClient,
			backupExectuor,
			kubeConfig,
			namespace,
		)
		parallelSkipper := run.NewParallelSkipper()
		return cron.NewExpressionCron(
			cronExpression,
			libsentry.NewSkipErrorAndReport(
				sentryClient,
				parallelSkipper.SkipParallel(backupAction.Run),
			),
		).Run(ctx)
	}
}

func CreateBackupAction(
	sentryClient libsentry.Client,
	backupExectuor BackupExectuor,
	kubeConfig string,
	namespace k8s.Namespace,
) run.Runnable {
	return NewBackupAction(
		sentryClient,
		NewK8sConnector(
			kubeConfig,
			namespace,
		),
		backupExectuor,
	)
}

func CreateSetupResourceDefinition(
	kubeConfig string,
	namespace k8s.Namespace,
	trigger run.Fire,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		k8sConnector := NewK8sConnector(
			kubeConfig,
			namespace,
		)
		if err := k8sConnector.SetupCustomResourceDefinition(ctx); err != nil {
			return errors.Wrap(ctx, err, "setup resource definition failed")
		}
		trigger.Fire()
		<-ctx.Done()
		return nil
	}
}

func CreateBackupExectuor(
	currentTimeGetter libtime.CurrentTimeGetter,
	backupRootDirectory Path,
	sshPrivateKey SSHPrivateKey,
) BackupExectuor {
	return NewBackupExectuorOnlyOnce(
		NewBackupExectuor(
			currentTimeGetter,
			NewRsyncExectuor(),
			backupRootDirectory,
			sshPrivateKey,
		),
	)
}
