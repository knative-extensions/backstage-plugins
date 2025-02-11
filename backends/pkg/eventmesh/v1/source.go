package v1

import (
	"encoding/json"
	"errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/eventing/pkg/apis/eventing"

	"knative.dev/backstage-plugins/backends/pkg/util"
)

// eventTypeEntry refers to an entry in the registry.knative.dev/eventTypes annotation.
type eventTypeEntry struct {
	Type        string `json:"type"`
	Schema      string `json:"schema,omitempty"`
	Description string `json:"description,omitempty"`
}

func convertSource(gvr schema.GroupVersionResource, crd unstructured.Unstructured, source *unstructured.Unstructured) (Source, error) {
	providedEventTypeTypes := []string{}

	crdAnnotations := crd.GetAnnotations()
	if eventTypesJson, ok := crdAnnotations[eventing.EventTypesAnnotationKey]; ok {
		var providedEventTypeEntries []eventTypeEntry
		if err := json.Unmarshal([]byte(eventTypesJson), &providedEventTypeEntries); err != nil {
			return Source{}, errors.New("failed to unmarshal event types")
		}

		providedEventTypeTypes = make([]string, len(providedEventTypeEntries))
		for i, entry := range providedEventTypeEntries {
			providedEventTypeTypes[i] = entry.Type
		}
	}

	src := Source{
		Namespace:              source.GetNamespace(),
		Name:                   source.GetName(),
		UID:                    string(source.GetUID()),
		Annotations:            util.FilterAnnotations(source.GetAnnotations()),
		Labels:                 source.GetLabels(),
		ProvidedEventTypeTypes: providedEventTypeTypes,
		// this field will be populated later on
		ProvidedEventTypes: []string{},
		Group:              gvr.Group,
		Kind:               source.GetKind(),
	}

	if sinkRef, ok := getSinkRef(source); ok {
		src.Sink = &sinkRef
	}

	return src, nil
}

func getSinkRef(u *unstructured.Unstructured) (GroupKindNamespacedName, bool) {
	stringMap, ok, err := unstructured.NestedStringMap(u.Object, "spec", "sink", "ref")
	if err != nil {
		return GroupKindNamespacedName{}, false
	}

	if !ok {
		return GroupKindNamespacedName{}, false
	}

	apiVersion, ok := stringMap["apiVersion"]
	if !ok {
		// if apiVersion is not present (e.g. using a URL as the sink), we don't care/
		// same story with others
		return GroupKindNamespacedName{}, false
	}
	kind, ok := stringMap["kind"]
	if !ok {
		return GroupKindNamespacedName{}, false
	}
	name, ok := stringMap["name"]
	if !ok {
		return GroupKindNamespacedName{}, false
	}

	return GroupKindNamespacedName{
		Group:     util.APIVersionToGroup(apiVersion),
		Kind:      kind,
		Namespace: u.GetNamespace(),
		Name:      name,
	}, true
}
