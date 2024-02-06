package eventmesh

import (
	"knative.dev/eventing/pkg/apis/eventing/v1beta2"
)

// EventType is a simplified representation of a Knative Eventing EventType that is easier to consume by the Backstage plugin.
type EventType struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	UID         string `json:"uid"`
	Description string `json:"description,omitempty"`
	SchemaData  string `json:"schemaData,omitempty"`
	SchemaURL   string `json:"schemaURL,omitempty"`
	// Labels of the event type. These are passed as is.
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations of the event type. These are passed as is, except that are filtered out by the FilterAnnotations function.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Reference is the ET's reference to a resource like a broker or a channel. It is in the format "<namespace>/<name>".
	Reference string `json:"reference,omitempty"`
	// ConsumedBy is a <namespace/name> list of the consumers of the event type.
	ConsumedBy []string `json:"consumedBy,omitempty"`
}

// NamespacedName returns the name and namespace of the event type in the format "<namespace>/<name>"
func (et EventType) NamespacedName() string {
	return NamespacedName(et.Namespace, et.Name)
}

// NamespacedType returns the type and namespace of the event type in the format "<namespace>/<type>"
func (et EventType) NamespacedType() string {
	return NamespacedName(et.Namespace, et.Type)
}

// convertEventType converts a Knative Eventing EventType to a simplified representation that is easier to consume by the Backstage plugin.
// see EventType.
func convertEventType(et *v1beta2.EventType) EventType {
	return EventType{
		Name:        et.Name,
		Namespace:   et.Namespace,
		Type:        et.Spec.Type,
		UID:         string(et.UID),
		Description: et.Spec.Description,
		SchemaData:  et.Spec.SchemaData,
		SchemaURL:   et.Spec.Schema.String(),
		Labels:      et.Labels,
		Annotations: FilterAnnotations(et.Annotations),
		Reference:   NamespacedRefName(et.Spec.Reference),
		// this field will be populated later on, when we have process the triggers
		ConsumedBy: make([]string, 0),
	}
}
