// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_kubernetes "k8s.io/client-go/kubernetes"
)

//counterfeiter:generate -o mocks/k8s-statefulset-deployer.go --fake-name K8sStatefulSetDeployer . StatefulSetDeployer
type StatefulSetDeployer interface {
	Deploy(ctx context.Context, statefulSet appsv1.StatefulSet) error
	Undeploy(ctx context.Context, namespace Namespace, name Name) error
}

func NewStatefulSetDeployer(
	clientset k8s_kubernetes.Interface,
) StatefulSetDeployer {
	return &statefulSetDeployer{
		clientset: clientset,
	}
}

type statefulSetDeployer struct {
	clientset k8s_kubernetes.Interface
}

func (s *statefulSetDeployer) Deploy(ctx context.Context, statefulSet appsv1.StatefulSet) error {
	_, err := s.clientset.AppsV1().StatefulSets(statefulSet.Namespace).Get(ctx, statefulSet.Name, metav1.GetOptions{})
	if err != nil {
		_, err = s.clientset.AppsV1().StatefulSets(statefulSet.Namespace).Create(ctx, &statefulSet, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(ctx, err, "create statefulSet failed")
		}
		glog.V(3).Infof("statefulSet %s created successful", statefulSet.Name)
		return nil
	}
	_, err = s.clientset.AppsV1().StatefulSets(statefulSet.Namespace).Update(ctx, &statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(ctx, err, "update statefulSet failed")
	}
	glog.V(3).Infof("statefulSet %s updated successful", statefulSet.Name)
	return nil

}

func (s *statefulSetDeployer) Undeploy(ctx context.Context, namespace Namespace, name Name) error {
	_, err := s.clientset.AppsV1().StatefulSets(namespace.String()).Get(ctx, name.String(), metav1.GetOptions{})
	if err != nil {
		glog.V(4).Infof("statefulSet '%s' not found => skip", name)
		return nil
	}
	if err := s.clientset.AppsV1().StatefulSets(namespace.String()).Delete(ctx, name.String(), metav1.DeleteOptions{}); err != nil {
		return err
	}
	glog.V(3).Infof("delete %s completed", name)
	return nil
}
