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

//counterfeiter:generate -o mocks/k8s-configmap-deployer.go --fake-name K8sConfigMapDeployer . ConfigMapDeployer
type ConfigMapDeployer interface {
	Get(ctx context.Context, namespace Namespace, name Name) (*v1.ConfigMap, error)
	Deploy(ctx context.Context, configmap v1.ConfigMap) error
	Undeploy(ctx context.Context, namespace Namespace, name Name) error
}

func NewConfigMapDeployer(
	clientset k8s_kubernetes.Interface,
) ConfigMapDeployer {
	return &configmapDeployer{
		clientset: clientset,
	}
}

type configmapDeployer struct {
	clientset k8s_kubernetes.Interface
}

func (s *configmapDeployer) Get(ctx context.Context, namespace Namespace, name Name) (*v1.ConfigMap, error) {
	cm, err := s.clientset.CoreV1().ConfigMaps(namespace.String()).Get(ctx, name.String(), metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(ctx, err, "get failed")
	}
	return cm, nil
}

func (s *configmapDeployer) Deploy(ctx context.Context, configmap v1.ConfigMap) error {
	currentConfigMap, err := s.clientset.CoreV1().ConfigMaps(configmap.Namespace).Get(ctx, configmap.Name, metav1.GetOptions{})
	if err != nil {
		_, err = s.clientset.CoreV1().ConfigMaps(configmap.Namespace).Create(ctx, &configmap, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(ctx, err, "create configmap failed")
		}
		glog.V(3).Infof("configmap %s created successful", configmap.Name)
		return nil
	}
	updateConfigMap := mergeConfigMap(*currentConfigMap, configmap)
	_, err = s.clientset.CoreV1().ConfigMaps(configmap.Namespace).Update(ctx, &updateConfigMap, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(ctx, err, "update configmap failed")
	}
	glog.V(3).Infof("configmap %s updated successful", configmap.Name)
	return nil
}

func (s *configmapDeployer) Undeploy(ctx context.Context, namespace Namespace, name Name) error {
	_, err := s.clientset.CoreV1().ConfigMaps(namespace.String()).Get(ctx, name.String(), metav1.GetOptions{})
	if err != nil {
		glog.V(4).Infof("configmap '%s' not found => skip", name)
		return nil
	}
	if err := s.clientset.CoreV1().ConfigMaps(namespace.String()).Delete(ctx, name.String(), metav1.DeleteOptions{}); err != nil {
		return err
	}
	glog.V(3).Infof("delete %s completed", name)
	return nil
}

func mergeConfigMap(current, new v1.ConfigMap) v1.ConfigMap {
	new.ObjectMeta.ResourceVersion = current.ObjectMeta.ResourceVersion
	return new
}
