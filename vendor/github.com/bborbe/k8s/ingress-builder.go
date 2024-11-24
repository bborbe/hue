// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//counterfeiter:generate -o mocks/k8s-ingress-builder.go --fake-name K8sIngressBuilder . IngressBuilder
type IngressBuilder interface {
	Build(ctx context.Context) (*v1.Ingress, error)
	SetObjectMetaBuilder(objectMetaBuilder ObjectMetaBuilder) IngressBuilder
	SetHost(host string) IngressBuilder
	SetServiceName(serviceName Name) IngressBuilder
	SetPath(path string) IngressBuilder
	SetIngressClassName(ingressClassName string) IngressBuilder
}

func NewIngressBuilder() IngressBuilder {
	return &ingressBuilder{
		serverPortName:   "http",
		ingressClassName: "traefik",
		pathType:         v1.PathTypePrefix,
		path:             "/",
	}
}

type ingressBuilder struct {
	serviceName       Name
	host              string
	ingressClassName  string
	serverPortName    string
	pathType          v1.PathType
	path              string
	objectMetaBuilder ObjectMetaBuilder
}

func (i *ingressBuilder) SetHost(host string) IngressBuilder {
	i.host = host
	return i
}

func (i *ingressBuilder) SetServiceName(serviceName Name) IngressBuilder {
	i.serviceName = serviceName
	return i
}

func (i *ingressBuilder) SetServerPortName(serverPortName string) IngressBuilder {
	i.serverPortName = serverPortName
	return i
}

func (i *ingressBuilder) SetPath(path string) IngressBuilder {
	i.path = path
	return i
}
func (i *ingressBuilder) SetPathType(pathType v1.PathType) IngressBuilder {
	i.pathType = pathType
	return i
}

func (i *ingressBuilder) SetObjectMetaBuilder(objectMetaBuilder ObjectMetaBuilder) IngressBuilder {
	i.objectMetaBuilder = objectMetaBuilder
	return i
}

func (i *ingressBuilder) SetIngressClassName(ingressClassName string) IngressBuilder {
	i.ingressClassName = ingressClassName
	return i
}

func (i *ingressBuilder) Validate(ctx context.Context) error {
	return validation.All{
		validation.Name("Host", validation.NotEmptyString(i.host)),
		validation.Name("IngressClassName", validation.NotEmptyString(i.ingressClassName)),
		validation.Name("Path", validation.NotEmptyString(i.path)),
		validation.Name("ServerPortName", validation.NotEmptyString(i.serverPortName)),
		validation.Name("ServiceName", i.serviceName),
		validation.Name("ObjectMetaBuilder", validation.NotNilAndValid(i.objectMetaBuilder)),
	}.Validate(ctx)
}

func (i *ingressBuilder) Build(ctx context.Context) (*v1.Ingress, error) {
	if err := i.Validate(ctx); err != nil {
		return nil, errors.Wrapf(ctx, err, "validate ingressBuilder failed")
	}

	objectMeta, err := i.objectMetaBuilder.Build(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "build objectMeta failed")
	}

	return &v1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: *objectMeta,
		Spec: v1.IngressSpec{
			IngressClassName: &i.ingressClassName,
			Rules: []v1.IngressRule{
				{
					Host: i.host,
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &i.pathType,
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: i.serviceName.String(),
											Port: v1.ServiceBackendPort{
												Name: i.serverPortName,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}
