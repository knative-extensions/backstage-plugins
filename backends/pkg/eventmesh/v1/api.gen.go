// Package v1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package v1

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// Broker Broker is a simplified representation of a Knative Eventing Broker that is easier to consume by the Backstage plugin.
type Broker struct {
	// Annotations Annotations of the broker.
	Annotations map[string]string `json:"annotations"`

	// Labels Labels of the broker.
	Labels map[string]string `json:"labels"`

	// Name Name of the broker.
	Name string `json:"name"`

	// Namespace Namespace of the broker.
	Namespace string `json:"namespace"`

	// ProvidedEventTypes List of event types provided by the broker.
	ProvidedEventTypes []string `json:"providedEventTypes"`

	// UID UID of the broker.
	UID string `json:"uid"`
}

// EventMesh EventMesh is the top-level struct that holds the event mesh data. It's the struct that's serialized and sent to the Backstage plugin.
type EventMesh struct {
	// Brokers Brokers is a list of all brokers in the cluster.
	Brokers []Broker `json:"brokers"`

	// EventTypes EventTypes is a list of all event types in the cluster. While we can embed the event types in the brokers, we keep them separate because not every event type is tied to a broker.
	EventTypes []EventType `json:"eventTypes"`

	// Sources Sources is a list of all sources in the cluster.
	Sources []Source `json:"sources"`

	// Subscribables Subscribables is a list of all subscribables in the cluster.
	Subscribables []Subscribable `json:"subscribables"`
}

// EventType EventType is a simplified representation of a Knative Eventing EventType that is easier to consume by the Backstage plugin.
type EventType struct {
	// Annotations Annotations of the event type. These are passed as is, except that are filtered out by the `FilterAnnotations` function.
	Annotations map[string]string `json:"annotations"`

	// ConsumedBy ConsumedBy is a `<namespace/name>` list of the consumers of the event type.
	ConsumedBy []string `json:"consumedBy"`

	// Description Description of the event type.
	Description *string `json:"description,omitempty"`

	// Labels Labels of the event type. These are passed as is.
	Labels map[string]string `json:"labels"`

	// Name Name of the event type.
	Name string `json:"name"`

	// Namespace Namespace of the event type.
	Namespace string `json:"namespace"`

	// Reference GroupKindNamespacedName is a struct that holds the group, kind, namespace, and name of a Kubernetes resource.
	Reference *GroupKindNamespacedName `json:"reference,omitempty"`

	// SchemaData Schema data.
	// Deprecated:
	SchemaData *string `json:"schemaData,omitempty"`

	// SchemaURL URL to the schema.
	SchemaURL *string `json:"schemaURL,omitempty"`

	// Type Type of the event.
	Type string `json:"type"`

	// Uid UID of the event type.
	Uid string `json:"uid"`
}

// GroupKindNamespacedName GroupKindNamespacedName is a struct that holds the group, kind, namespace, and name of a Kubernetes resource.
type GroupKindNamespacedName struct {
	// Group Kubernetes API group of the resource, without the version.
	Group string `json:"group"`

	// Kind Kubernetes API kind of the resource.
	Kind string `json:"kind"`

	// Name Name of the resource.
	Name string `json:"name"`

	// Namespace Namespace of the resource.
	Namespace string `json:"namespace"`
}

// Source Source is a simplified representation of a Knative Eventing Source that is easier to consume by the Backstage plugin.
type Source struct {
	// Annotations Annotations of the source.
	Annotations map[string]string `json:"annotations"`

	// Group Kubernetes API group of the source, without the version.
	Group string `json:"group"`

	// Kind Kubernetes API kind of the source.
	Kind string `json:"kind"`

	// Labels Labels of the source.
	Labels map[string]string `json:"labels"`

	// Name Name of the source.
	Name string `json:"name"`

	// Namespace Namespace of the source.
	Namespace string `json:"namespace"`

	// ProvidedEventTypeTypes List of EventType types provided by the source. These are simply the `spec.type` of the EventTypes.
	ProvidedEventTypeTypes []string `json:"providedEventTypeTypes"`

	// ProvidedEventTypes List of EventTypes provided by the source. These are the `<namespace/name>` of the EventTypes.
	ProvidedEventTypes []string `json:"providedEventTypes"`

	// Sink GroupKindNamespacedName is a struct that holds the group, kind, namespace, and name of a Kubernetes resource.
	Sink *GroupKindNamespacedName `json:"sink,omitempty"`

	// UID UID of the source.
	UID string `json:"uid"`
}

// Subscribable Subscribable is a simplified representation of a Knative Eventing Subscribable that is easier to consume by the Backstage plugin. These subscribables can be channels at the moment.
type Subscribable struct {
	// Annotations Annotations of the subscribable.
	Annotations map[string]string `json:"annotations"`

	// Group Kubernetes API group of the subscribable, without the version.
	Group string `json:"group"`

	// Kind Kubernetes API kind of the subscribable.
	Kind string `json:"kind"`

	// Labels Labels of the subscribable.
	Labels map[string]string `json:"labels"`

	// Name Name of the subscribable.
	Name string `json:"name"`

	// Namespace Namespace of the subscribable.
	Namespace string `json:"namespace"`

	// ProvidedEventTypes List of event types provided by the subscribable.
	ProvidedEventTypes []string `json:"providedEventTypes"`

	// UID UID of the subscribable.
	UID string `json:"uid"`
}
