package eventmesh

import (
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

// Broker is a simplified representation of a Knative Eventing Broker that is easier to consume by the Backstage plugin.
type Broker struct {
	// Namespace of the broker
	Namespace string `json:"namespace"`

	// Name of the broker
	Name string `json:"name"`

	// UID of the broker
	UID string `json:"uid"`

	// Labels of the broker. These are passed as is.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations of the broker. These are passed as is, except that are filtered out by the FilterAnnotations function.
	Annotations map[string]string `json:"annotations,omitempty"`

	// ProvidedEventTypes is a list of event types that the broker provides.
	// This is a list of strings, where each string is a "<namespace>/<name>" of the event type.
	ProvidedEventTypes []string `json:"providedEventTypes,omitempty"`
}

// GetNamespacedName returns the name and namespace of the broker in the format "<namespace>/<name>"
func (b Broker) GetNamespacedName() string {
	return NamespacedName(b.Namespace, b.Name)
}

// convertBroker converts a Knative Eventing Broker to a simplified representation that is easier to consume by the Backstage plugin.
// see Broker.
func convertBroker(br *eventingv1.Broker) Broker {
	return Broker{
		Name:        br.Name,
		Namespace:   br.Namespace,
		UID:         string(br.UID),
		Labels:      br.Labels,
		Annotations: FilterAnnotations(br.Annotations),
		// this field will be populated later on, when we have the list of event types
		ProvidedEventTypes: []string{},
	}
}
