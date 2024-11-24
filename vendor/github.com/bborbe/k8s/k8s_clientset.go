// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"
	k8s_kubernetes "k8s.io/client-go/kubernetes"
	k8s_rest "k8s.io/client-go/rest"
	k8s_clientcmd "k8s.io/client-go/tools/clientcmd"
)

func CreateClientset(kubeconfig string) (k8s_kubernetes.Interface, error) {
	config, err := CreateConfig(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "create k8s config failed")
	}
	return k8s_kubernetes.NewForConfig(config)
}

func CreateConfig(kubeconfig string) (*k8s_rest.Config, error) {
	if len(kubeconfig) > 0 {
		glog.V(4).Infof("create kube config from flags")
		return k8s_clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	glog.V(4).Infof("create in cluster kube config")
	return k8s_rest.InClusterConfig()
}
