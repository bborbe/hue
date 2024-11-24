package pkg

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/golang/glog"
)

type RsyncExectuor interface {
	Rsync(ctx context.Context, args ...string) error
}

func NewRsyncExectuor() RsyncExectuor {
	return &rsyncExectuor{}
}

type rsyncExectuor struct {
}

func (r *rsyncExectuor) Rsync(ctx context.Context, args ...string) error {
	glog.V(2).Infof("run: rsync %s", strings.Join(args, " "))
	cmd := exec.Command("rsync", args...)
	if glog.V(2) {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	glog.V(2).Infof("rsync started")
	if err := cmd.Wait(); err != nil {
		var msg *exec.ExitError
		if errors.As(err, &msg) {
			glog.V(2).Infof("rsync closed with exit error")
			if waitstatus, ok := msg.Sys().(syscall.WaitStatus); ok {
				glog.V(2).Infof("rsync closed with exit error: %d", waitstatus.ExitStatus())
				if waitstatus.ExitStatus() == 24 {
					glog.V(2).Infof("rsync finished with vanished file error")
					return nil
				}
			}
		}
		return err
	}
	glog.V(2).Infof("rsync finished")
	return nil
}
