package v1

import (
	"knative.dev/backstage-plugins/backends/pkg/util"
	"knative.dev/eventing/pkg/apis/eventing/v1beta2"
)

// NamespacedName returns the name and namespace of the event type in the format "<namespace>/<name>"
func (et EventType) NamespacedName() string {
	return util.NamespacedName(et.Namespace, et.Name)
}

// NamespacedType returns the type and namespace of the event type in the format "<namespace>/<type>"
func (et EventType) NamespacedType() string {
	return util.NamespacedName(et.Namespace, et.Type)
}

// convertEventType converts a Knative Eventing EventType to a simplified representation that is easier to consume by the Backstage plugin.
// see EventType.
func convertEventType(et *v1beta2.EventType) EventType {
	var reference *GroupKindNamespacedName
	if et.Spec.Reference != nil {
		reference = &GroupKindNamespacedName{
			Group:     util.APIVersionToGroup(et.Spec.Reference.APIVersion),
			Kind:      et.Spec.Reference.Kind,
			Namespace: et.Namespace,
			Name:      et.Spec.Reference.Name,
		}
	}
	return EventType{
		Name:        et.Name,
		Namespace:   et.Namespace,
		Type:        et.Spec.Type,
		Uid:         string(et.UID),
		Description: util.ToStrPtrOrNil(et.Spec.Description),
		SchemaData:  util.ToStrPtrOrNil(et.Spec.SchemaData),
		SchemaURL:   util.ToStrPtrOrNil(et.Spec.Schema.String()),
		Labels:      et.Labels,
		Annotations: util.FilterAnnotations(et.Annotations),
		Reference:   reference,
		// this field will be populated later on, when we have process the triggers
		ConsumedBy: make([]string, 0),
	}
}
