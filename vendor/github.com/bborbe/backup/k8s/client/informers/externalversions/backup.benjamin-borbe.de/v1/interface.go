// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	internalinterfaces "github.com/bborbe/backup/k8s/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Targets returns a TargetInformer.
	Targets() TargetInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// Targets returns a TargetInformer.
func (v *version) Targets() TargetInformer {
	return &targetInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}