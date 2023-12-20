package eventmesh

import (
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

type Broker struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	UID         string            `json:"uid"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	//
	ProvidedEventTypes []string `json:"providedEventTypes,omitempty"`
}

func (b Broker) GetNameAndNamespace() string {
	return NameAndNamespace(b.Namespace, b.Name)
}

func convertBroker(br *eventingv1.Broker) Broker {
	return Broker{
		Name:        br.Name,
		Namespace:   br.Namespace,
		UID:         string(br.UID),
		Labels:      br.Labels,
		Annotations: FilterAnnotations(br.Annotations),
		// to be filled later
		ProvidedEventTypes: []string{},
	}
}
