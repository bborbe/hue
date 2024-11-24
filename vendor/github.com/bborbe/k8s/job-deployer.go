// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var foreground = metav1.DeletePropagationForeground

//counterfeiter:generate -o mocks/k8s-job-deployer.go --fake-name K8sJobDeployer . JobDeployer
type JobDeployer interface {
	Deploy(ctx context.Context, job batchv1.Job) error
	Undeploy(ctx context.Context, namespace Namespace, name Name) error
}

func NewJobDeployer(
	clientset kubernetes.Interface,
) JobDeployer {
	return &jobDeployer{
		clientset: clientset,
	}
}

type jobDeployer struct {
	clientset kubernetes.Interface
}

func (s *jobDeployer) Deploy(ctx context.Context, job batchv1.Job) error {
	glog.V(3).Infof("deploy %s started", job.Name)
	if err := s.Undeploy(ctx, Namespace(job.Namespace), Name(job.Name)); err != nil {
		return errors.Wrap(ctx, err, "undeploy failed")
	}
	if _, err := s.clientset.BatchV1().Jobs(job.Namespace).Create(ctx, &job, metav1.CreateOptions{}); err != nil {
		return errors.Wrap(ctx, err, "create job failed")
	}
	glog.V(3).Infof("job %s created successful", job.Name)
	return nil
}

func (s *jobDeployer) Undeploy(ctx context.Context, namespace Namespace, name Name) error {
	glog.V(3).Infof("delete %s started", name)
	if _, err := s.clientset.BatchV1().Jobs(namespace.String()).Get(ctx, name.String(), metav1.GetOptions{}); err != nil {
		glog.V(4).Infof("job '%s' not found => skip", name)
		return nil
	}
	if err := s.clientset.BatchV1().Jobs(namespace.String()).Delete(ctx, name.String(), metav1.DeleteOptions{
		PropagationPolicy: &foreground,
	}); err != nil {
		return err
	}
	glog.V(3).Infof("delete %s completed", name)
	return nil
}
