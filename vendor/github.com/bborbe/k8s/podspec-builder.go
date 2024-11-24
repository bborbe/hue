// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//counterfeiter:generate -o mocks/k8s-podspec-builder.go --fake-name K8sPodSpecBuilder . PodSpecBuilder
type PodSpecBuilder interface {
	Build(ctx context.Context) (*corev1.PodSpec, error)
	SetContainers(containers []corev1.Container) PodSpecBuilder
	SetVolumes(volumes []corev1.Volume) PodSpecBuilder
	SetRestartPolicy(restartPolicy corev1.RestartPolicy) PodSpecBuilder
	SetAffinity(affinity corev1.Affinity) PodSpecBuilder
}

func NewPodSpecBuilder() PodSpecBuilder {
	return &podSpecBuilder{
		restartPolicy: corev1.RestartPolicyAlways,
	}
}

type podSpecBuilder struct {
	name          string
	objectMeta    metav1.ObjectMeta
	containers    []corev1.Container
	volumes       []corev1.Volume
	restartPolicy corev1.RestartPolicy
	affinity      *corev1.Affinity
}

func (p *podSpecBuilder) SetRestartPolicy(restartPolicy corev1.RestartPolicy) PodSpecBuilder {
	p.restartPolicy = restartPolicy
	return p
}

func (p *podSpecBuilder) SetAffinity(affinity corev1.Affinity) PodSpecBuilder {
	p.affinity = &affinity
	return p
}

func (p *podSpecBuilder) SetContainers(containers []corev1.Container) PodSpecBuilder {
	p.containers = containers
	return p
}

func (p *podSpecBuilder) SetVolumes(volumes []corev1.Volume) PodSpecBuilder {
	p.volumes = volumes
	return p
}

func (p *podSpecBuilder) Validate(ctx context.Context) error {
	return validation.All{}.Validate(ctx)
}

func (p *podSpecBuilder) Build(ctx context.Context) (*corev1.PodSpec, error) {
	if err := p.Validate(ctx); err != nil {
		return nil, errors.Wrapf(ctx, err, "validate ingressBuilder failed")
	}
	return &corev1.PodSpec{
		Volumes:       p.volumes,
		Containers:    p.containers,
		RestartPolicy: p.restartPolicy,
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: "docker",
			},
		},
		Affinity: p.affinity,
	}, nil
}
