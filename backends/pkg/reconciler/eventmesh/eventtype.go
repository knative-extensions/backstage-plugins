package eventmesh

import (
	"knative.dev/eventing/pkg/apis/eventing/v1beta2"
)

type EventType struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	UID         string            `json:"uid"`
	Description string            `json:"description,omitempty"`
	SchemaData  string            `json:"schemaData,omitempty"`
	SchemaURL   string            `json:"schemaURL,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

func convertEventType(et *v1beta2.EventType) EventType {
	// TODO: more information!
	return EventType{
		Name:        et.Name,
		Namespace:   et.Namespace,
		Type:        et.Spec.Type,
		UID:         string(et.UID),
		Description: et.Spec.Description,
		SchemaData:  et.Spec.SchemaData,
		SchemaURL:   et.Spec.Schema.String(),
		Labels:      et.Labels,
		Annotations: filterAnnotations(et.Annotations),
	}
}
