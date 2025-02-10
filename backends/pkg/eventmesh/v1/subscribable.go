package v1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/backstage-plugins/backends/pkg/util"
)

func convertSubscribable(gvr schema.GroupVersionResource, u *unstructured.Unstructured) Subscribable {
	return Subscribable{
		Namespace:   u.GetNamespace(),
		Name:        u.GetName(),
		UID:         string(u.GetUID()),
		Annotations: util.FilterAnnotations(u.GetAnnotations()),
		Labels:      u.GetLabels(),
		Group:       gvr.Group,
		Kind:        u.GetKind(),
		// this field will be populated later on
		ProvidedEventTypes: []string{},
	}
}
