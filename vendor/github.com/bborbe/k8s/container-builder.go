// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

//counterfeiter:generate -o mocks/k8s-container-builder.go --fake-name K8sContainerBuilder . ContainerBuilder
type ContainerBuilder interface {
	Build(ctx context.Context) (*corev1.Container, error)
	Validate(ctx context.Context) error
	SetEnvBuilder(envBuilder EnvBuilder) ContainerBuilder
	SetImage(image string) ContainerBuilder
	SetName(name Name) ContainerBuilder
	SetCommand(command []string) ContainerBuilder
	SetArgs(args []string) ContainerBuilder
	SetPorts(ports []corev1.ContainerPort) ContainerBuilder
	SetVolumeMounts(volumeMounts []corev1.VolumeMount) ContainerBuilder
	AddVolumeMounts(volumeMounts ...corev1.VolumeMount) ContainerBuilder
	SetCpuLimit(cpuLimit string) ContainerBuilder
	SetCpuRequest(cpuRequest string) ContainerBuilder
	SetMemoryLimit(memoryLimit string) ContainerBuilder
	SetMemoryRequest(memoryRequest string) ContainerBuilder
	SetLivenessProbe(livenessProbe corev1.Probe) ContainerBuilder
	SetReadinessProbe(readinessProbe corev1.Probe) ContainerBuilder
	SetRestartPolicy(restartPolicy corev1.ContainerRestartPolicy) ContainerBuilder
}

func NewContainerBuilder() ContainerBuilder {
	return &containerBuilder{
		cpuLimit:      "50m",
		cpuRequest:    "20m",
		memoryLimit:   "50Mi",
		memoryRequest: "20Mi",
		envBuilder:    NewEnvBuilder(),
	}
}

type containerBuilder struct {
	envBuilder     EnvBuilder
	name           Name
	image          string
	args           []string
	command        []string
	ports          []corev1.ContainerPort
	volumeMounts   []corev1.VolumeMount
	cpuLimit       string
	cpuRequest     string
	memoryLimit    string
	memoryRequest  string
	livenessProbe  *corev1.Probe
	readinessProbe *corev1.Probe
	restartPolicy  *corev1.ContainerRestartPolicy
}

func (c *containerBuilder) SetRestartPolicy(restartPolicy corev1.ContainerRestartPolicy) ContainerBuilder {
	c.restartPolicy = &restartPolicy
	return c
}

func (c *containerBuilder) SetLivenessProbe(livenessProbe corev1.Probe) ContainerBuilder {
	c.livenessProbe = &livenessProbe
	return c
}

func (c *containerBuilder) SetReadinessProbe(readinessProbe corev1.Probe) ContainerBuilder {
	c.readinessProbe = &readinessProbe
	return c
}

func (c *containerBuilder) AddVolumeMounts(volumeMounts ...corev1.VolumeMount) ContainerBuilder {
	c.volumeMounts = append(c.volumeMounts, volumeMounts...)
	return c
}

func (c *containerBuilder) SetCpuLimit(cpuLimit string) ContainerBuilder {
	c.cpuLimit = cpuLimit
	return c
}

func (c *containerBuilder) SetCpuRequest(cpuRequest string) ContainerBuilder {
	c.cpuRequest = cpuRequest
	return c
}

func (c *containerBuilder) SetMemoryLimit(memoryLimit string) ContainerBuilder {
	c.memoryLimit = memoryLimit
	return c
}

func (c *containerBuilder) SetMemoryRequest(memoryRequest string) ContainerBuilder {
	c.memoryRequest = memoryRequest
	return c
}

func (c *containerBuilder) SetVolumeMounts(volumeMounts []corev1.VolumeMount) ContainerBuilder {
	c.volumeMounts = volumeMounts
	return c
}

func (c *containerBuilder) SetPorts(ports []corev1.ContainerPort) ContainerBuilder {
	c.ports = ports
	return c
}

func (c *containerBuilder) SetCommand(command []string) ContainerBuilder {
	c.command = command
	return c
}

func (c *containerBuilder) SetArgs(args []string) ContainerBuilder {
	c.args = args
	return c
}

func (c *containerBuilder) SetEnvBuilder(envBuilder EnvBuilder) ContainerBuilder {
	c.envBuilder = envBuilder
	return c
}

func (c *containerBuilder) SetName(name Name) ContainerBuilder {
	c.name = name
	return c
}

func (c *containerBuilder) SetImage(image string) ContainerBuilder {
	c.image = image
	return c
}

func (c *containerBuilder) Validate(ctx context.Context) error {
	return validation.All{
		validation.Name("Name", validation.NotEmptyString(c.name)),
		validation.Name("EnvBuilder", validation.NotNilAndValid(c.envBuilder)),
	}.Validate(ctx)
}

func (c *containerBuilder) Build(ctx context.Context) (*corev1.Container, error) {
	if err := c.Validate(ctx); err != nil {
		return nil, errors.Wrapf(ctx, err, "validate containerBuilder failed")
	}

	envVars, err := c.envBuilder.Build(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "build env failed")
	}

	return &corev1.Container{
		Name:    c.name.String(),
		Image:   c.image,
		Command: c.command,
		Args:    c.args,
		Ports:   c.ports,
		Env:     envVars,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"cpu":    resource.MustParse(c.cpuLimit),
				"memory": resource.MustParse(c.memoryLimit),
			},
			Requests: corev1.ResourceList{
				"cpu":    resource.MustParse(c.cpuRequest),
				"memory": resource.MustParse(c.memoryRequest),
			},
		},
		VolumeMounts:    c.volumeMounts,
		ImagePullPolicy: corev1.PullAlways,
		LivenessProbe:   c.livenessProbe,
		ReadinessProbe:  c.readinessProbe,
		RestartPolicy:   c.restartPolicy,
	}, nil
}
