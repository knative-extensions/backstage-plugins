package v1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/utils/ptr"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/injection/clients/dynamicclient"

	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventingv1beta2 "knative.dev/eventing/pkg/apis/eventing/v1beta2"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"

	fakeclientset "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	testingv1 "knative.dev/eventing/pkg/reconciler/testing/v1"
	testingv1beta2 "knative.dev/eventing/pkg/reconciler/testing/v1beta2"
)

func TestBuildEventMesh(t *testing.T) {
	tests := []struct {
		name         string
		brokers      []*eventingv1.Broker
		eventTypes   []*eventingv1beta2.EventType
		triggers     []*eventingv1.Trigger
		extraObjects []runtime.Object
		want         EventMesh
		error        bool
	}{
		{
			name: "With 1 broker, 1 type, 1 trigger",
			brokers: []*eventingv1.Broker{
				testingv1.NewBroker("test-broker", "test-ns",
					// following fields are not used in any logic and simply returned
					WithBrokerUID("test-broker-uid"),
					WithBrokerLabels(map[string]string{"test-broker-label": "foo"}),
					WithBrokerAnnotations(map[string]string{"test-broker-annotation": "foo"}),
				),
			},
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
					// following fields are not used in any logic and simply returned
					testingv1beta2.WithEventTypeDescription("test-eventtype-description"),
					WithEventTypeUID("test-eventtype-uid"),
					WithEventTypeSchema(&apis.URL{
						Scheme: "http",
						Host:   "test-eventtype-schema",
					}),
					WithEventTypeSchemaData("test-eventtype-schema-data"),
					testingv1beta2.WithEventTypeLabels(map[string]string{"test-eventtype-label": "foo"}),
					WithEventTypeAnnotations(map[string]string{"test-eventtype-annotation": "foo"}),
				),
			},
			triggers: []*eventingv1.Trigger{
				testingv1.NewTrigger("test-trigger", "test-ns", "test-broker",
					testingv1.WithTriggerSubscriberRef(
						metav1.GroupVersionKind{
							Group:   "",
							Version: "v1",
							Kind:    "Service",
						},
						"test-subscriber",
						"test-ns",
					),
					WithEventTypeFilter("test-eventtype-type"),
				),
			},
			extraObjects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-subscriber",
						Namespace: "test-ns",
						Labels:    map[string]string{"backstage.io/kubernetes-id": "test-subscriber"},
					},
				},
			},
			want: EventMesh{
				Brokers: []Broker{
					{
						Name:               "test-broker",
						Namespace:          "test-ns",
						UID:                "test-broker-uid",
						Labels:             map[string]string{"test-broker-label": "foo"},
						Annotations:        map[string]string{"test-broker-annotation": "foo"},
						ProvidedEventTypes: []string{"test-ns/test-eventtype"}},
				},
				EventTypes: []EventType{
					{
						Name:        "test-eventtype",
						Namespace:   "test-ns",
						Type:        "test-eventtype-type",
						Uid:         "test-eventtype-uid",
						Description: ptr.To("test-eventtype-description"),
						SchemaData:  ptr.To("test-eventtype-schema-data"),
						SchemaURL:   ptr.To("http://test-eventtype-schema"),
						Labels:      map[string]string{"test-eventtype-label": "foo"},
						Annotations: map[string]string{"test-eventtype-annotation": "foo"},
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{"test-subscriber"},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources:       make([]Source, 0),
			},
		},
		{
			name: "With 1 broker, 1 type, no triggers",
			brokers: []*eventingv1.Broker{
				testingv1.NewBroker("test-broker", "test-ns"),
			},
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
				),
			},
			want: EventMesh{
				Brokers: []Broker{
					{
						Name:               "test-broker",
						Namespace:          "test-ns",
						ProvidedEventTypes: []string{"test-ns/test-eventtype"}},
				},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype",
						Namespace: "test-ns",
						Type:      "test-eventtype-type",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources:       make([]Source, 0),
			},
		},
		{
			name: "With 1 broker and 2 eventtypes with different spec.types, no triggers",
			brokers: []*eventingv1.Broker{
				testingv1.NewBroker("test-broker", "test-ns"),
			},
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype-1", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type-1"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
				),
				testingv1beta2.NewEventType("test-eventtype-2", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type-2"),
				),
			},
			want: EventMesh{
				Brokers: []Broker{
					{
						Name:               "test-broker",
						Namespace:          "test-ns",
						ProvidedEventTypes: []string{"test-ns/test-eventtype-1"}},
				},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype-1",
						Namespace: "test-ns",
						Type:      "test-eventtype-type-1",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{},
					},
					{
						Name:       "test-eventtype-2",
						Namespace:  "test-ns",
						Type:       "test-eventtype-type-2",
						Reference:  nil,
						ConsumedBy: []string{},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources:       make([]Source, 0),
			},
		},
		{
			name: "With 2 brokers and 2 eventtypes with same spec.types, no triggers",
			brokers: []*eventingv1.Broker{
				testingv1.NewBroker("test-broker-1", "test-ns"),
				testingv1.NewBroker("test-broker-2", "test-ns"),
			},
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype-1", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker-1", "test-ns")),
				),
				testingv1beta2.NewEventType("test-eventtype-2", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker-2", "test-ns")),
				),
			},
			want: EventMesh{
				Brokers: []Broker{
					{
						Name:               "test-broker-1",
						Namespace:          "test-ns",
						ProvidedEventTypes: []string{"test-ns/test-eventtype-1"},
					},
					{
						Name:               "test-broker-2",
						Namespace:          "test-ns",
						ProvidedEventTypes: []string{"test-ns/test-eventtype-2"},
					},
				},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype-1",
						Namespace: "test-ns",
						Type:      "test-eventtype-type",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker-1",
						},
						ConsumedBy: []string{},
					},
					{
						Name:      "test-eventtype-2",
						Namespace: "test-ns",
						Type:      "test-eventtype-type",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker-2",
						},
						ConsumedBy: []string{},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources:       make([]Source, 0),
			},
		},
		{
			name: "Ignore triggers that are not bound to a broker that exists",
			brokers: []*eventingv1.Broker{
				testingv1.NewBroker("test-broker", "test-ns"),
			},
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
				),
			},
			triggers: []*eventingv1.Trigger{
				testingv1.NewTrigger("test-trigger", "test-ns", "UNKNOWN-BROKER",
					testingv1.WithTriggerSubscriberRef(
						metav1.GroupVersionKind{
							Group:   "",
							Version: "v1",
							Kind:    "Service",
						},
						"test-subscriber",
						"test-ns",
					),
					WithEventTypeFilter("test-eventtype-type"),
				),
			},
			extraObjects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-subscriber",
						Namespace: "test-ns",
						Labels:    map[string]string{"backstage.io/kubernetes-id": "test-subscriber"},
					},
				},
			},
			want: EventMesh{
				Brokers: []Broker{
					{
						Name:               "test-broker",
						Namespace:          "test-ns",
						ProvidedEventTypes: []string{"test-ns/test-eventtype"}},
				},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype",
						Namespace: "test-ns",
						Type:      "test-eventtype-type",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources:       make([]Source, 0),
			},
		},
		{
			name: "Ignore triggers that are not bound to a subscriber that exists",
			brokers: []*eventingv1.Broker{
				testingv1.NewBroker("test-broker", "test-ns"),
			},
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
				),
			},
			triggers: []*eventingv1.Trigger{
				testingv1.NewTrigger("test-trigger", "test-ns", "test-broker",
					testingv1.WithTriggerSubscriberRef(
						metav1.GroupVersionKind{
							Group:   "",
							Version: "v1",
							Kind:    "Service",
						},
						"test-subscriber",
						"test-ns",
					),
					WithEventTypeFilter("test-eventtype-type"),
				),
			},
			extraObjects: []runtime.Object{},
			want: EventMesh{
				Brokers: []Broker{
					{
						Name:               "test-broker",
						Namespace:          "test-ns",
						ProvidedEventTypes: []string{"test-ns/test-eventtype"}},
				},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype",
						Namespace: "test-ns",
						Type:      "test-eventtype-type",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources:       make([]Source, 0),
			},
		},
		{
			name: "Ignore triggers that are bound to a subscriber that is not registered on Backstage",
			brokers: []*eventingv1.Broker{
				testingv1.NewBroker("test-broker", "test-ns"),
			},
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
				),
			},
			triggers: []*eventingv1.Trigger{
				testingv1.NewTrigger("test-trigger", "test-ns", "test-broker",
					testingv1.WithTriggerSubscriberRef(
						metav1.GroupVersionKind{
							Group:   "",
							Version: "v1",
							Kind:    "Service",
						},
						"test-subscriber",
						"test-ns",
					),
					WithEventTypeFilter("test-eventtype-type"),
				),
			},
			extraObjects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-subscriber",
						Namespace: "test-ns",
						Labels:    map[string]string{"NO": "NO"},
					},
				},
			},
			want: EventMesh{
				Brokers: []Broker{
					{
						Name:               "test-broker",
						Namespace:          "test-ns",
						ProvidedEventTypes: []string{"test-ns/test-eventtype"}},
				},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype",
						Namespace: "test-ns",
						Type:      "test-eventtype-type",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources:       make([]Source, 0),
			},
		},
		{
			name: "Trigger with no filter subscribes to all event types provided by the broker",
			brokers: []*eventingv1.Broker{
				testingv1.NewBroker("test-broker", "test-ns"),
			},
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype-1", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type-1"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
				),
				testingv1beta2.NewEventType("test-eventtype-2", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type-2"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
				),
			},
			triggers: []*eventingv1.Trigger{
				testingv1.NewTrigger("test-trigger", "test-ns", "test-broker",
					testingv1.WithTriggerSubscriberRef(
						metav1.GroupVersionKind{
							Group:   "",
							Version: "v1",
							Kind:    "Service",
						},
						"test-subscriber",
						"test-ns",
					),
				),
			},
			extraObjects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-subscriber",
						Namespace: "test-ns",
						Labels:    map[string]string{"backstage.io/kubernetes-id": "test-subscriber"},
					},
				},
			},
			want: EventMesh{
				Brokers: []Broker{
					{
						Name:      "test-broker",
						Namespace: "test-ns",
						ProvidedEventTypes: []string{
							"test-ns/test-eventtype-1",
							"test-ns/test-eventtype-2",
						}},
				},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype-1",
						Namespace: "test-ns",
						Type:      "test-eventtype-type-1",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{"test-subscriber"},
					},
					{
						Name:      "test-eventtype-2",
						Namespace: "test-ns",
						Type:      "test-eventtype-type-2",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{"test-subscriber"},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources:       make([]Source, 0),
			},
		},
		{
			name: "Trigger has an accept all types filter subscribes to all event types provided by the broker",
			brokers: []*eventingv1.Broker{
				testingv1.NewBroker("test-broker", "test-ns"),
			},
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype-1", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type-1"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
				),
				testingv1beta2.NewEventType("test-eventtype-2", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type-2"),
					testingv1beta2.WithEventTypeReference(brokerReference("test-broker", "test-ns")),
				),
			},
			triggers: []*eventingv1.Trigger{
				testingv1.NewTrigger("test-trigger", "test-ns", "test-broker",
					testingv1.WithTriggerSubscriberRef(
						metav1.GroupVersionKind{
							Group:   "",
							Version: "v1",
							Kind:    "Service",
						},
						"test-subscriber",
						"test-ns",
					),
					WithEventTypeFilter(""),
				),
			},
			extraObjects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-subscriber",
						Namespace: "test-ns",
						Labels:    map[string]string{"backstage.io/kubernetes-id": "test-subscriber"},
					},
				},
			},
			want: EventMesh{
				Brokers: []Broker{
					{
						Name:      "test-broker",
						Namespace: "test-ns",
						ProvidedEventTypes: []string{
							"test-ns/test-eventtype-1",
							"test-ns/test-eventtype-2",
						}},
				},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype-1",
						Namespace: "test-ns",
						Type:      "test-eventtype-type-1",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{"test-subscriber"},
					},
					{
						Name:      "test-eventtype-2",
						Namespace: "test-ns",
						Type:      "test-eventtype-type-2",
						Reference: &GroupKindNamespacedName{
							Group:     "eventing.knative.dev",
							Kind:      "Broker",
							Namespace: "test-ns",
							Name:      "test-broker",
						},
						ConsumedBy: []string{"test-subscriber"},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources:       make([]Source, 0),
			},
		},
		{
			name: "With 1 channel, 1 type",
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
					testingv1beta2.WithEventTypeReference(reference("messaging.knative.dev/v1", "InMemoryChannel", "test-ns", "test-imc")),
				),
			},
			triggers: []*eventingv1.Trigger{},
			extraObjects: []runtime.Object{
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "inmemorychannels.messaging.knative.dev",
						Labels: map[string]string{
							"messaging.knative.dev/subscribable": "true",
						},
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
						Group: "messaging.knative.dev",
						Names: apiextensionsv1.CustomResourceDefinitionNames{
							Kind:     "InMemoryChannel",
							ListKind: "InMemoryChannelList",
							Plural:   "inmemorychannels",
						},
						Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
							{
								Name:    "v1",
								Served:  true,
								Storage: true,
							},
						},
					},
				},
				&messagingv1.InMemoryChannel{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-imc",
						Namespace: "test-ns",
					},
				},
			},
			want: EventMesh{
				Brokers: []Broker{},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype",
						Namespace: "test-ns",
						Type:      "test-eventtype-type",
						Reference: &GroupKindNamespacedName{
							Group:     "messaging.knative.dev",
							Kind:      "InMemoryChannel",
							Namespace: "test-ns",
							Name:      "test-imc",
						},
						ConsumedBy: []string{},
					},
				},
				Subscribables: []Subscribable{
					{
						Group:              "messaging.knative.dev",
						Kind:               "InMemoryChannel",
						Name:               "test-imc",
						Namespace:          "test-ns",
						ProvidedEventTypes: []string{"test-ns/test-eventtype"},
					},
				},
				Sources: make([]Source, 0),
			},
		},
		{
			name: "With 1 channel, 1 type, 1 subscriber",
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
					testingv1beta2.WithEventTypeReference(reference("messaging.knative.dev/v1", "InMemoryChannel", "test-ns", "test-imc")),
				),
			},
			triggers: []*eventingv1.Trigger{},
			extraObjects: []runtime.Object{
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "inmemorychannels.messaging.knative.dev",
						Labels: map[string]string{
							"messaging.knative.dev/subscribable": "true",
						},
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
						Group: "messaging.knative.dev",
						Names: apiextensionsv1.CustomResourceDefinitionNames{
							Kind:     "InMemoryChannel",
							ListKind: "InMemoryChannelList",
							Plural:   "inmemorychannels",
						},
						Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
							{
								Name:    "v1",
								Served:  true,
								Storage: true,
							},
						},
					},
				},
				&messagingv1.InMemoryChannel{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-imc",
						Namespace: "test-ns",
					},
				},
			},
			want: EventMesh{
				Brokers: []Broker{},
				EventTypes: []EventType{
					{
						Name:      "test-eventtype",
						Namespace: "test-ns",
						Type:      "test-eventtype-type",
						Reference: &GroupKindNamespacedName{
							Group:     "messaging.knative.dev",
							Kind:      "InMemoryChannel",
							Namespace: "test-ns",
							Name:      "test-imc",
						},
						ConsumedBy: []string{},
					},
				},
				Subscribables: []Subscribable{
					{
						Group:              "messaging.knative.dev",
						Kind:               "InMemoryChannel",
						Name:               "test-imc",
						Namespace:          "test-ns",
						ProvidedEventTypes: []string{"test-ns/test-eventtype"},
					},
				},
				Sources: make([]Source, 0),
			},
		},
		{
			name: "With 1 source, 2 types",
			eventTypes: []*eventingv1beta2.EventType{
				testingv1beta2.NewEventType("test-eventtype", "test-ns",
					testingv1beta2.WithEventTypeType("test-eventtype-type"),
				),
				testingv1beta2.NewEventType("test-eventtype-used-in-source", "test-ns",
					testingv1beta2.WithEventTypeType("dev.knative.apiserver.resource.add"),
				),
			},
			triggers: []*eventingv1.Trigger{},
			extraObjects: []runtime.Object{
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sources.knative.dev",
						Labels: map[string]string{
							"duck.knative.dev/source": "true",
						},
						Annotations: map[string]string{
							"registry.knative.dev/eventTypes": `
							[
							  {
							    "type": "dev.knative.apiserver.resource.add",
							    "description": "CloudEvent type used for add operations when in Resource mode"
							  }
							]
							`,
						},
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
						Group: "sources.knative.dev",
						Names: apiextensionsv1.CustomResourceDefinitionNames{
							Kind:     "ApiServerSource",
							ListKind: "ApiServerSourceList",
							Plural:   "apiserversources",
						},
						Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
							{
								Name:    "v1",
								Served:  true,
								Storage: true,
							},
						},
					},
				},
				&sourcesv1.ApiServerSource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-src",
						Namespace: "test-ns",
					},
				},
			},
			want: EventMesh{
				Brokers: []Broker{},
				EventTypes: []EventType{
					{
						Name:       "test-eventtype",
						Namespace:  "test-ns",
						Type:       "test-eventtype-type",
						ConsumedBy: []string{},
					},
					{
						Name:       "test-eventtype-used-in-source",
						Namespace:  "test-ns",
						Type:       "dev.knative.apiserver.resource.add",
						ConsumedBy: []string{},
					},
				},
				Subscribables: make([]Subscribable, 0),
				Sources: []Source{
					{
						Group:                  "sources.knative.dev",
						Kind:                   "ApiServerSource",
						Name:                   "test-src",
						Namespace:              "test-ns",
						ProvidedEventTypeTypes: []string{"dev.knative.apiserver.resource.add"},
						ProvidedEventTypes:     []string{"test-ns/test-eventtype-used-in-source"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		logger := zap.NewNop().Sugar()
		v1beta2objects := make([]runtime.Object, 0, 20)
		for _, et := range tt.eventTypes {
			v1beta2objects = append(v1beta2objects, et)
		}

		for _, b := range tt.brokers {
			v1beta2objects = append(v1beta2objects, b)
		}

		for _, t := range tt.triggers {
			v1beta2objects = append(v1beta2objects, t)
		}
		sc := runtime.NewScheme()
		_ = corev1.AddToScheme(sc)
		_ = eventingv1.AddToScheme(sc)
		_ = messagingv1.AddToScheme(sc)
		_ = sourcesv1.AddToScheme(sc)
		_ = apiextensionsv1.AddToScheme(sc)

		fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(sc, tt.extraObjects...)

		ctx := context.TODO()
		ctx = context.WithValue(ctx, dynamicclient.Key{}, fakeDynamicClient)

		fakeClient := fakeclientset.NewSimpleClientset(v1beta2objects...)

		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildEventMesh(ctx, fakeClient, fakeDynamicClient, logger)
			if (err != nil) != tt.error {
				t.Errorf("BuildEventMesh() error = %v, error %v", err, tt.error)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("BuildEventMesh() (-want, +got):", diff)
			}
		})
	}
}

func WithEventTypeFilter(et string) testingv1.TriggerOption {
	return func(a *eventingv1.Trigger) {
		if a.Spec.Filter == nil {
			a.Spec.Filter = &eventingv1.TriggerFilter{}
		}
		if a.Spec.Filter.Attributes == nil {
			a.Spec.Filter.Attributes = make(map[string]string)
		}
		a.Spec.Filter.Attributes["type"] = et
	}
}

func brokerReference(brokerName, namespace string) *duckv1.KReference {
	return reference("eventing.knative.dev/v1", "Broker", namespace, brokerName)
}

func reference(apiVersion string, kind string, namespace string, name string) *duckv1.KReference {
	return &duckv1.KReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Namespace:  namespace,
		Name:       name,
	}
}

func WithEventTypeUID(uid string) testingv1beta2.EventTypeOption {
	return func(a *eventingv1beta2.EventType) {
		a.UID = types.UID(uid)
	}
}

func WithEventTypeSchema(url *apis.URL) testingv1beta2.EventTypeOption {
	return func(a *eventingv1beta2.EventType) {
		a.Spec.Schema = url
	}
}

func WithEventTypeSchemaData(d string) testingv1beta2.EventTypeOption {
	return func(a *eventingv1beta2.EventType) {
		a.Spec.SchemaData = d
	}
}

func WithEventTypeAnnotations(annotations map[string]string) testingv1beta2.EventTypeOption {
	return func(a *eventingv1beta2.EventType) {
		a.ObjectMeta.Annotations = annotations
	}
}

func WithBrokerUID(uid string) testingv1.BrokerOption {
	return func(a *eventingv1.Broker) {
		a.UID = types.UID(uid)
	}
}

func WithBrokerAnnotations(annotations map[string]string) testingv1.BrokerOption {
	return func(a *eventingv1.Broker) {
		a.ObjectMeta.Annotations = annotations
	}
}

func WithBrokerLabels(labels map[string]string) testingv1.BrokerOption {
	return func(a *eventingv1.Broker) {
		a.ObjectMeta.Labels = labels
	}
}
