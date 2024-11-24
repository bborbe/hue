// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//counterfeiter:generate -o mocks/k8s-deployment-builder.go --fake-name K8sDeploymentBuilder . DeploymentBuilder
type DeploymentBuilder interface {
	Build(ctx context.Context) (*appsv1.Deployment, error)
	SetObjectMetaBuilder(objectMetaBuilder ObjectMetaBuilder) DeploymentBuilder
	SetContainersBuilder(containersBuilder ContainersBuilder) DeploymentBuilder
	SetName(name Name) DeploymentBuilder
	SetReplicas(replicas int32) DeploymentBuilder
	SetComponent(component string) DeploymentBuilder
	SetServiceAccountName(serviceAccountName string) DeploymentBuilder
	AddVolumes(volumes ...corev1.Volume) DeploymentBuilder
	SetVolumes(volumes []corev1.Volume) DeploymentBuilder
	SetAffinity(affinity corev1.Affinity) DeploymentBuilder
}

func NewDeploymentBuilder() DeploymentBuilder {
	return &deploymentBuilder{
		replicas: 1,
	}
}

type deploymentBuilder struct {
	component          string
	name               Name
	objectMetaBuilder  ObjectMetaBuilder
	replicas           int32
	serviceAccountName string
	volumes            []corev1.Volume
	containersBuilder  ContainersBuilder
	affinity           *corev1.Affinity
}

func (d *deploymentBuilder) SetAffinity(affinity corev1.Affinity) DeploymentBuilder {
	d.affinity = &affinity
	return d
}

func (d *deploymentBuilder) SetContainersBuilder(containersBuilder ContainersBuilder) DeploymentBuilder {
	d.containersBuilder = containersBuilder
	return d
}

func (d *deploymentBuilder) AddVolumes(volumes ...corev1.Volume) DeploymentBuilder {
	d.volumes = append(d.volumes, volumes...)
	return d
}

func (d *deploymentBuilder) SetVolumes(volumes []corev1.Volume) DeploymentBuilder {
	d.volumes = volumes
	return d
}

func (d *deploymentBuilder) SetServiceAccountName(serviceAccountName string) DeploymentBuilder {
	d.serviceAccountName = serviceAccountName
	return d
}

func (d *deploymentBuilder) SetObjectMetaBuilder(objectMetaBuilder ObjectMetaBuilder) DeploymentBuilder {
	d.objectMetaBuilder = objectMetaBuilder
	return d
}

func (d *deploymentBuilder) SetName(name Name) DeploymentBuilder {
	d.name = name
	return d
}

func (d *deploymentBuilder) SetReplicas(replicas int32) DeploymentBuilder {
	d.replicas = replicas
	return d
}

func (d *deploymentBuilder) SetComponent(component string) DeploymentBuilder {
	d.component = component
	return d
}

func (d *deploymentBuilder) Validate(ctx context.Context) error {
	return validation.All{
		validation.Name("ObjectMeta", validation.NotNilAndValid(d.objectMetaBuilder)),
		validation.Name("ContainersBuilder", validation.NotNilAndValid(d.containersBuilder)),
	}.Validate(ctx)
}

func (d *deploymentBuilder) Build(ctx context.Context) (*appsv1.Deployment, error) {
	if err := d.Validate(ctx); err != nil {
		return nil, errors.Wrapf(ctx, err, "validate deploymentBuilder failed")
	}

	objectMeta, err := d.objectMetaBuilder.Build(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "build objectMeta failed")
	}

	containers, err := d.containersBuilder.Build(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "build containers failed")
	}

	maxUnavailable := intstr.FromInt32(1)
	maxSurge := intstr.FromInt32(1)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: *objectMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas: &d.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": d.name.String(),
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &maxUnavailable,
					MaxSurge:       &maxSurge,
				},
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "9090",
						"prometheus.io/scheme": "http",
						"prometheus.io/scrape": "true",
					},
					Labels: map[string]string{
						"component": d.component,
						"app":       d.name.String(),
					},
				},
				Spec: corev1.PodSpec{
					Affinity:           d.affinity,
					Containers:         containers,
					ServiceAccountName: d.serviceAccountName,
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "docker",
						},
					},
					Volumes: d.volumes,
				},
			},
		},
	}, nil
}
