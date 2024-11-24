// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/bborbe/collection"
	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//counterfeiter:generate -o mocks/k8s-job-builder.go --fake-name K8sJobBuilder . JobBuilder
type JobBuilder interface {
	Build(ctx context.Context) (*batchv1.Job, error)
	SetName(name Name) JobBuilder
	SetObjectMeta(objectMeta metav1.ObjectMeta) JobBuilder
	SetComponent(component string) JobBuilder
	AddLabel(key, value string) JobBuilder
	SetLabels(labels map[string]string) JobBuilder
	SetPodSpec(podSpec corev1.PodSpec) JobBuilder
}

func NewJobBuilder() JobBuilder {
	return &jobBuilder{
		labels: map[string]string{},
	}
}

type jobBuilder struct {
	name       Name
	objectMeta metav1.ObjectMeta
	component  string
	labels     map[string]string
	podSpec    corev1.PodSpec
}

func (j *jobBuilder) SetPodSpec(podSpec corev1.PodSpec) JobBuilder {
	j.podSpec = podSpec
	return j
}

func (j *jobBuilder) SetObjectMeta(objectMeta metav1.ObjectMeta) JobBuilder {
	j.objectMeta = objectMeta
	return j
}

func (j *jobBuilder) SetName(name Name) JobBuilder {
	j.name = name
	return j
}

func (j *jobBuilder) SetComponent(component string) JobBuilder {
	return j.AddLabel("component", component)
}

func (j *jobBuilder) SetLabels(labels map[string]string) JobBuilder {
	j.labels = labels
	return j
}

func (j *jobBuilder) AddLabel(key, value string) JobBuilder {
	j.labels[key] = value
	return j
}

func (j *jobBuilder) Validate(ctx context.Context) error {
	return validation.All{
		validation.Name("ObjectMeta", validation.NotNil(j.objectMeta)),
	}.Validate(ctx)
}

func (j *jobBuilder) Build(ctx context.Context) (*batchv1.Job, error) {
	if err := j.Validate(ctx); err != nil {
		return nil, errors.Wrapf(ctx, err, "validate jobBuilder failed")
	}

	j.AddLabel("app", j.name.String())

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: j.objectMeta,
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Labels:      j.labels,
				},
				Spec: j.podSpec,
			},
			TTLSecondsAfterFinished: collection.Ptr(int32(600)),
			BackoffLimit:            collection.Ptr(int32(4)),
			CompletionMode:          collection.Ptr(batchv1.NonIndexedCompletion),
			Completions:             collection.Ptr(int32(1)),
			Parallelism:             collection.Ptr(int32(1)),
			PodReplacementPolicy:    collection.Ptr(batchv1.TerminatingOrFailed),
		},
	}, nil
}
