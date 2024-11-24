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

//counterfeiter:generate -o mocks/k8s-service-builder.go --fake-name K8sServiceBuilder . ServiceBuilder
type ServiceBuilder interface {
	Build(ctx context.Context) (*corev1.Service, error)
	SetObjectMetaBuilder(objectMetaBuilder ObjectMetaBuilder) ServiceBuilder
	SetName(name Name) ServiceBuilder
	SetServicePortName(servicePortName string) ServiceBuilder
	SetServicePortNumber(servicePortNumber int32) ServiceBuilder
}

func NewServiceBuilder() ServiceBuilder {
	return &serviceBuilder{
		servicePortName:   "http",
		servicePortNumber: 9090,
	}
}

type serviceBuilder struct {
	name              Name
	objectMetaBuilder ObjectMetaBuilder
	servicePortNumber int32
	servicePortName   string
}

func (s *serviceBuilder) SetObjectMetaBuilder(objectMetaBuilder ObjectMetaBuilder) ServiceBuilder {
	s.objectMetaBuilder = objectMetaBuilder
	return s
}

func (s *serviceBuilder) SetName(name Name) ServiceBuilder {
	s.name = name
	return s
}

func (s *serviceBuilder) SetServicePortNumber(servicePortNumber int32) ServiceBuilder {
	s.servicePortNumber = servicePortNumber
	return s
}

func (s *serviceBuilder) SetServicePortName(servicePortName string) ServiceBuilder {
	s.servicePortName = servicePortName
	return s
}

func (s *serviceBuilder) Validate(ctx context.Context) error {
	return validation.All{
		validation.Name("Name", validation.NotEmptyString(s.name)),
		validation.Name("ObjectMetaBuilder", validation.NotNilAndValid(s.objectMetaBuilder)),
	}.Validate(ctx)
}

func (s *serviceBuilder) Build(ctx context.Context) (*corev1.Service, error) {
	if err := s.Validate(ctx); err != nil {
		return nil, errors.Wrapf(ctx, err, "validate serviceBuilder failed")
	}

	objectMeta, err := s.objectMetaBuilder.Build(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "build objectMeta failed")
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: *objectMeta,
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: s.servicePortName,
					Port: s.servicePortNumber,
				},
			},
			Selector: map[string]string{
				"app": s.name.String(),
			},
		},
	}, nil
}
