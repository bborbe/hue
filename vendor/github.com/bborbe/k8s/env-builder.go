// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/validation"
	corev1 "k8s.io/api/core/v1"
)

//counterfeiter:generate -o mocks/k8s-env-builder.go --fake-name K8sEnvBuilder . EnvBuilder
type EnvBuilder interface {
	Add(name, value string) EnvBuilder
	AddSecret(name, secret, key string) EnvBuilder
	Build(ctx context.Context) ([]corev1.EnvVar, error)
	AddFieldRef(name string, apiVersion string, fieldPath string) EnvBuilder
	Validate(ctx context.Context) error
}

func NewEnvBuilder() EnvBuilder {
	return &envBuilder{
		envs: make([]corev1.EnvVar, 0),
	}
}

type envBuilder struct {
	envs []corev1.EnvVar
}

func (e *envBuilder) Validate(ctx context.Context) error {
	return validation.All{}.Validate(ctx)
}

func (e *envBuilder) AddFieldRef(name string, apiVersion string, fieldPath string) EnvBuilder {
	e.envs = append(e.envs, corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: apiVersion,
				FieldPath:  fieldPath,
			},
		},
	})
	return e
}

func (e *envBuilder) Add(name, value string) EnvBuilder {
	e.envs = append(e.envs, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
	return e
}

func (e *envBuilder) AddSecret(name, secret, key string) EnvBuilder {
	e.envs = append(e.envs, corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret,
				},
				Key: key,
			},
		},
	})
	return e
}

func (e *envBuilder) Build(ctx context.Context) ([]corev1.EnvVar, error) {
	return e.envs, nil
}
