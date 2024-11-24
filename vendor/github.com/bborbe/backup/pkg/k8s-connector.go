// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"time"

	"github.com/bborbe/errors"
	"github.com/bborbe/k8s"
	"github.com/golang/glog"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsClient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	backupv1 "github.com/bborbe/backup/k8s/apis/backup.benjamin-borbe.de/v1"
	"github.com/bborbe/backup/k8s/client/clientset/versioned"
	"github.com/bborbe/backup/k8s/client/informers/externalversions"
)

const (
	defaultResync = 5 * time.Minute
	name          = "targets.backup.benjamin-borbe.de"
)

//counterfeiter:generate -o ../mocks/k8s-connector.go --fake-name K8sConnector . K8sConnector
type K8sConnector interface {
	SetupCustomResourceDefinition(ctx context.Context) error
	Listen(ctx context.Context, resourceEventHandler cache.ResourceEventHandler) error
	Targets(ctx context.Context) (backupv1.Targets, error)
	Target(ctx context.Context, name string) (*backupv1.Target, error)
}

func NewK8sConnector(
	kubeconfig string,
	namespace k8s.Namespace,
) K8sConnector {
	return &k8sConnector{
		kubeconfig: kubeconfig,
		namespace:  namespace,
	}
}

type k8sConnector struct {
	kubeconfig string
	namespace  k8s.Namespace
}

func (k *k8sConnector) Target(ctx context.Context, name string) (*backupv1.Target, error) {
	config, err := k.createKubernetesConfig()
	if err != nil {
		return nil, errors.Wrap(ctx, err, "build k8s config failed")
	}
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "build clientset failed")
	}
	target, err := clientset.BackupV1().Targets(k.namespace.String()).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(ctx, err, "list target failed")
	}
	return target, nil
}

func (k *k8sConnector) Targets(ctx context.Context) (backupv1.Targets, error) {
	config, err := k.createKubernetesConfig()
	if err != nil {
		return nil, errors.Wrap(ctx, err, "build k8s config failed")
	}
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "build clientset failed")
	}
	targetList, err := clientset.BackupV1().Targets(k.namespace.String()).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(ctx, err, "list target failed")
	}
	return targetList.Items, nil
}

func (k *k8sConnector) Listen(
	ctx context.Context,
	resourceEventHandler cache.ResourceEventHandler,
) error {
	config, err := k.createKubernetesConfig()
	if err != nil {
		return errors.Wrap(ctx, err, "build k8s config failed")
	}
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return errors.Wrap(ctx, err, "build clientset failed")
	}
	informerFactory := externalversions.NewSharedInformerFactory(clientset, defaultResync)
	_, err = informerFactory.
		Backup().
		V1().
		Targets().
		Informer().
		AddEventHandler(resourceEventHandler)
	if err != nil {
		return errors.Wrap(ctx, err, "add event handler failed")
	}

	stopCh := make(chan struct{})
	glog.V(2).Infof("listen for events")
	informerFactory.Start(stopCh)
	select {
	case <-ctx.Done():
		glog.V(0).Infof("listen canceled")
	case <-stopCh:
		glog.V(0).Infof("listen stopped")
	}
	return nil
}

func (k *k8sConnector) SetupCustomResourceDefinition(ctx context.Context) error {
	config, err := k.createKubernetesConfig()
	if err != nil {
		return errors.Wrap(ctx, err, "build k8s config failed")
	}
	clientset, err := apiextensionsClient.NewForConfig(config)
	if err != nil {
		return errors.Wrap(ctx, err, "build clientset failed")
	}
	customResourceDefinition, err := clientset.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		glog.V(2).Infof("CustomResourceDefinition '%s' not found (%v) => create", name, err)
		if err := k.createCrd(ctx, clientset); err != nil {
			return errors.Wrap(ctx, err, "create crd failed")
		}
		return nil
	}
	if err := k.updateCrd(ctx, customResourceDefinition, clientset); err != nil {
		return errors.Wrap(ctx, err, "create crd failed")
	}
	return nil
}

func (k *k8sConnector) updateCrd(ctx context.Context, customResourceDefinition *v1.CustomResourceDefinition, clientset *apiextensionsClient.Clientset) error {
	customResourceDefinition.Spec = createSpec()
	if _, err := clientset.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, customResourceDefinition, metav1.UpdateOptions{}); err != nil {
		return errors.Wrap(ctx, err, "update CustomResourceDefinition failed")
	}
	glog.V(2).Infof("CustomResourceDefinitions '%s' updated", name)
	return nil
}

func (k *k8sConnector) createCrd(ctx context.Context, clientset *apiextensionsClient.Clientset) error {
	_, err := clientset.ApiextensionsV1().CustomResourceDefinitions().Create(
		ctx,
		&v1.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apiextensions.k8s.io/v1",
				Kind:       "CustomResourceDefinition",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: createSpec(),
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		return errors.Wrap(ctx, err, "create CustomResourceDefinition failed")
	}
	glog.V(2).Infof("CustomResourceDefinition '%s' created", name)
	return nil
}

func (k *k8sConnector) createKubernetesConfig() (*rest.Config, error) {
	if len(k.kubeconfig) > 0 {
		glog.V(3).Infof("create kube config from flags")
		return clientcmd.BuildConfigFromFlags("", k.kubeconfig)
	}
	glog.V(3).Infof("create in cluster kube config")
	return rest.InClusterConfig()
}

func boolPointer(value bool) *bool {
	return &value
}

func createSpec() v1.CustomResourceDefinitionSpec {
	return v1.CustomResourceDefinitionSpec{
		Group: "backup.benjamin-borbe.de",
		Names: v1.CustomResourceDefinitionNames{
			Kind:     "Target",
			ListKind: "TargetList",
			Plural:   "targets",
			Singular: "target",
		},
		Scope: "Namespaced",
		Versions: []v1.CustomResourceDefinitionVersion{
			{
				Name:    "v1",
				Served:  true,
				Storage: true,
				Schema: &v1.CustomResourceValidation{
					OpenAPIV3Schema: &v1.JSONSchemaProps{
						XPreserveUnknownFields: boolPointer(true),
					},
				},
			},
		},
	}
}
