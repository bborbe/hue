// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"
	"reflect"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_kubernetes "k8s.io/client-go/kubernetes"
)

//counterfeiter:generate -o mocks/k8s-ingress-deployer.go --fake-name K8sIngressDeployer . IngressDeployer
type IngressDeployer interface {
	Deploy(ctx context.Context, ingress v1.Ingress) error
	Undeploy(ctx context.Context, namespace Namespace, name Name) error
}

func NewIngressDeployer(
	clientset k8s_kubernetes.Interface,
) IngressDeployer {
	return &ingressDeployer{
		clientset: clientset,
	}
}

type ingressDeployer struct {
	clientset k8s_kubernetes.Interface
}

func (s *ingressDeployer) Deploy(ctx context.Context, ingress v1.Ingress) error {
	currentIngress, err := s.clientset.NetworkingV1().Ingresses(ingress.Namespace).Get(ctx, ingress.Name, metav1.GetOptions{})
	if err != nil {
		_, err = s.clientset.NetworkingV1().Ingresses(ingress.Namespace).Create(ctx, &ingress, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(ctx, err, "create ingress failed")
		}
		glog.V(3).Infof("ingress %s created successful", ingress.Name)
		return nil
	}
	if IngressEqual(ingress, *currentIngress) {
		glog.V(3).Infof("ingress %s already update to date => skip", ingress.Name)
		return nil
	}
	_, err = s.clientset.NetworkingV1().Ingresses(ingress.Namespace).Update(ctx, &ingress, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(ctx, err, "update ingress failed")
	}
	glog.V(3).Infof("ingress %s updated successful", ingress.Name)
	return nil

}

func (s *ingressDeployer) Undeploy(ctx context.Context, namespace Namespace, name Name) error {
	_, err := s.clientset.NetworkingV1().Ingresses(namespace.String()).Get(ctx, name.String(), metav1.GetOptions{})
	if err != nil {
		glog.V(4).Infof("ingress '%s' not found => skip", name)
		return nil
	}
	if err := s.clientset.NetworkingV1().Ingresses(namespace.String()).Delete(ctx, name.String(), metav1.DeleteOptions{}); err != nil {
		return err
	}
	glog.V(3).Infof("delete %s completed", name)
	return nil
}

func IngressEqual(a, b v1.Ingress) bool {
	if reflect.DeepEqual(a.ObjectMeta.Labels, b.ObjectMeta.Labels) == false {
		glog.V(3).Infof("ObjectMeta.Labels not equal %#v %#v", a.ObjectMeta.Labels, b.ObjectMeta.Labels)
		return false
	}
	if reflect.DeepEqual(a.ObjectMeta.Annotations, b.ObjectMeta.Annotations) == false {
		glog.V(3).Infof("ObjectMeta.Annotations not equal %#v %#v", a.ObjectMeta.Annotations, b.ObjectMeta.Annotations)
		return false
	}
	if reflect.DeepEqual(a.Spec, b.Spec) == false {
		glog.V(3).Infof("Spec not equal %#v %#v", a.Spec, b.Spec)
		return false
	}
	return true
}
