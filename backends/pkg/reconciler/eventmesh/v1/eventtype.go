package v1

import (
	"knative.dev/eventing/pkg/apis/eventing/v1beta2"
	"knative.dev/eventing/pkg/apis/eventing/v1beta3"

	"knative.dev/backstage-plugins/backends/pkg/util"
)

// NamespacedName returns the name and namespace of the event type in the format "<namespace>/<name>"
func (et EventType) NamespacedName() string {
	return util.NamespacedName(et.Namespace, et.Name)
}

// NamespacedType returns the type and namespace of the event type in the format "<namespace>/<type>"
func (et EventType) NamespacedType() string {
	return util.NamespacedName(et.Namespace, et.Type)
}

// TODO: remove
// convertEventType converts a Knative Eventing EventType to a simplified representation that is easier to consume by the Backstage plugin.
// see EventType.
func convertEventType(et *v1beta2.EventType) EventType {
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
		Reference:   util.ToStrPtrOrNil(util.NamespacedRefName(et.Spec.Reference)),
		// this field will be populated later on, when we have process the triggers
		ConsumedBy: make([]string, 0),
	}
}

// convertEventType converts a Knative Eventing EventType to a simplified representation that is easier to consume by the Backstage plugin.
// see EventType.
func convertEventTypev1beta3(et *v1beta3.EventType) EventType {
	cet := EventType{
		Name:        et.Name,
		Namespace:   et.Namespace,
		Uid:         string(et.UID),
		Description: util.ToStrPtrOrNil(et.Spec.Description),
		Labels:      et.Labels,
		Annotations: util.FilterAnnotations(et.Annotations),
		Reference:   util.ToStrPtrOrNil(util.NamespacedRefName(et.Spec.Reference)),
		// this field will be populated later on, when we have process the triggers
		ConsumedBy: make([]string, 0),
	}

	if len(et.Spec.Attributes) == 0 {
		return cet
	}

	for _, attr := range et.Spec.Attributes {
		switch attr.Name {
		case "type": // TODO: any CE constant for these?
			cet.Type = attr.Value
		case "schemadata":
			cet.SchemaURL = util.ToStrPtrOrNil(attr.Value)
		}
	}

	return cet
}
