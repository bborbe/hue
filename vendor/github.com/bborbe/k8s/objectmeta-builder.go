// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//counterfeiter:generate -o mocks/k8s-objectmeta-builder.go --fake-name K8sObjectMetaBuilder . ObjectMetaBuilder
type ObjectMetaBuilder interface {
	Build(ctx context.Context) (*metav1.ObjectMeta, error)
	SetGenerateName(generateName string) ObjectMetaBuilder
	SetName(name Name) ObjectMetaBuilder
	SetNamespace(namespace Namespace) ObjectMetaBuilder
	SetComponent(component string) ObjectMetaBuilder
	AddLabel(key, value string) ObjectMetaBuilder
	AddAnnotation(key, value string) ObjectMetaBuilder
	SetFinalizers(finalizers []string) ObjectMetaBuilder
	Validate(ctx context.Context) error
}

func NewObjectMetaBuilder() ObjectMetaBuilder {
	return &objectMetaBuilder{
		labels:      map[string]string{},
		annotations: map[string]string{},
		finalizers:  []string{},
	}
}

type objectMetaBuilder struct {
	component    string
	name         Name
	namespace    Namespace
	annotations  map[string]string
	labels       map[string]string
	generateName string
	finalizers   []string
}

func (o *objectMetaBuilder) SetFinalizers(finalizers []string) ObjectMetaBuilder {
	o.finalizers = finalizers
	return o
}

func (o *objectMetaBuilder) AddLabel(key, value string) ObjectMetaBuilder {
	o.labels[key] = value
	return o
}

func (o *objectMetaBuilder) AddAnnotation(key, value string) ObjectMetaBuilder {
	o.annotations[key] = value
	return o
}

func (o *objectMetaBuilder) SetGenerateName(generateName string) ObjectMetaBuilder {
	o.generateName = generateName
	return o
}

func (o *objectMetaBuilder) SetName(name Name) ObjectMetaBuilder {
	o.name = name
	return o
}

func (o *objectMetaBuilder) SetNamespace(namespace Namespace) ObjectMetaBuilder {
	o.namespace = namespace
	return o
}

func (o *objectMetaBuilder) SetComponent(component string) ObjectMetaBuilder {
	o.component = component
	o.labels["component"] = o.component
	return o
}

func (o *objectMetaBuilder) Validate(ctx context.Context) error {
	return validation.All{
		validation.Name("Name", o.name),
		validation.Name("Namespace", validation.NotEmptyString(o.namespace)),
	}.Validate(ctx)
}

func (o *objectMetaBuilder) Build(ctx context.Context) (*metav1.ObjectMeta, error) {
	if err := o.Validate(ctx); err != nil {
		return nil, errors.Wrapf(ctx, err, "validate objectMetaBuilder failed")
	}
	return &metav1.ObjectMeta{
		Name:         o.name.String(),
		GenerateName: o.generateName,
		Namespace:    o.namespace.String(),
		Labels:       o.labels,
		Annotations:  o.annotations,
		Finalizers:   o.finalizers,
	}, nil
}
