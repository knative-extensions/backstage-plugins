package v1

import (
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"

	"knative.dev/backstage-plugins/backends/pkg/util"
)

// GetNamespacedName returns the name and namespace of the broker in the format "<namespace>/<name>"
func (b Broker) GetNamespacedName() string {
	return util.NamespacedName(b.Namespace, b.Name)
}

// convertBroker converts a Knative Eventing Broker to a simplified representation that is easier to consume by the Backstage plugin.
// see Broker.
func convertBroker(br *eventingv1.Broker) Broker {
	return Broker{
		Name:        br.Name,
		Namespace:   br.Namespace,
		UID:         string(br.UID),
		Labels:      br.Labels,
		Annotations: util.FilterAnnotations(br.Annotations),
		// this field will be populated later on, when we have the list of event types
		ProvidedEventTypes: []string{},
	}
}
