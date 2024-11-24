// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
	corev1 "k8s.io/api/core/v1"
)

//counterfeiter:generate -o mocks/k8s-container-builder.go --fake-name K8sContainerBuilder . ContainerBuilder
type ContainersBuilder interface {
	Build(ctx context.Context) ([]corev1.Container, error)
	AddContainerBuilder(containerBuilder ContainerBuilder) ContainersBuilder
	SetContainerBuilder(containerBuilders []ContainerBuilder) ContainersBuilder
	Validate(ctx context.Context) error
}

func NewContainersBuilder() ContainersBuilder {
	return &containersBuilder{}
}

type containersBuilder struct {
	containerBuilders []ContainerBuilder
}

func (c *containersBuilder) AddContainerBuilder(containerBuilder ContainerBuilder) ContainersBuilder {
	c.containerBuilders = append(c.containerBuilders, containerBuilder)
	return c
}

func (c *containersBuilder) SetContainerBuilder(containerBuilders []ContainerBuilder) ContainersBuilder {
	c.containerBuilders = containerBuilders
	return c
}

func (c *containersBuilder) Validate(ctx context.Context) error {
	return validation.All{
		validation.Name("ContainerBuilders", validation.NotEmptySlice(c.containerBuilders)),
	}.Validate(ctx)
}

func (c *containersBuilder) Build(ctx context.Context) ([]corev1.Container, error) {
	if err := c.Validate(ctx); err != nil {
		return nil, errors.Wrapf(ctx, err, "validate containersBuilder failed")
	}
	var result []corev1.Container
	for _, containerBuilder := range c.containerBuilders {
		container, err := containerBuilder.Build(ctx)
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "build container failed")
		}
		result = append(result, *container)
	}
	return result, nil
}
