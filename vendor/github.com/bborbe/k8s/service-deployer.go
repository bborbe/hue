// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_kubernetes "k8s.io/client-go/kubernetes"
)

//counterfeiter:generate -o mocks/k8s-service-deployer.go --fake-name K8sServiceDeployer . ServiceDeployer
type ServiceDeployer interface {
	Deploy(ctx context.Context, service v1.Service) error
	Undeploy(ctx context.Context, namespace Namespace, name Name) error
}

func NewServiceDeployer(
	clientset k8s_kubernetes.Interface,
) ServiceDeployer {
	return &serviceDeployer{
		clientset: clientset,
	}
}

type serviceDeployer struct {
	clientset k8s_kubernetes.Interface
}

func (s *serviceDeployer) Deploy(ctx context.Context, service v1.Service) error {
	currentService, err := s.clientset.CoreV1().Services(service.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil {
		_, err = s.clientset.CoreV1().Services(service.Namespace).Create(ctx, &service, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(ctx, err, "create service failed")
		}
		glog.V(3).Infof("service %s created successful", service.Name)
		return nil
	}
	updateService := mergeService(*currentService, service)
	_, err = s.clientset.CoreV1().Services(service.Namespace).Update(ctx, &updateService, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(ctx, err, "update service failed")
	}
	glog.V(3).Infof("service %s updated successful", service.Name)
	return nil
}

func (s *serviceDeployer) Undeploy(ctx context.Context, namespace Namespace, name Name) error {
	_, err := s.clientset.CoreV1().Services(namespace.String()).Get(ctx, name.String(), metav1.GetOptions{})
	if err != nil {
		glog.V(4).Infof("service '%s' not found => skip", name)
		return nil
	}
	if err := s.clientset.CoreV1().Services(namespace.String()).Delete(ctx, name.String(), metav1.DeleteOptions{}); err != nil {
		return err
	}
	glog.V(3).Infof("delete %s completed", name)
	return nil
}

func mergeService(current, new v1.Service) v1.Service {
	new.Spec.ClusterIP = current.Spec.ClusterIP
	new.ObjectMeta.ResourceVersion = current.ObjectMeta.ResourceVersion
	return new
}
