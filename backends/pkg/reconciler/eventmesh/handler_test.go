package eventmesh

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"testing"

	"go.uber.org/zap"

	duckv1 "knative.dev/pkg/apis/duck/v1"

	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventingv1beta2 "knative.dev/eventing/pkg/apis/eventing/v1beta2"

	reconcilertestingv1 "knative.dev/eventing/pkg/reconciler/testing/v1"
	reconcilertestingv1beta2 "knative.dev/eventing/pkg/reconciler/testing/v1beta2"

	testingv1 "knative.dev/eventing/pkg/reconciler/testing/v1"
	testingv1beta2 "knative.dev/eventing/pkg/reconciler/testing/v1beta2"
)

func TestBuildEventMesh(t *testing.T) {
	tests := []struct {
		name       string
		brokers    []*eventingv1.Broker
		eventTypes []*eventingv1beta2.EventType
		want       EventMesh
		error      bool
	}{
		{
			name: "With 1 broker and 1 type",
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
			want: EventMesh{
				Brokers: []*Broker{
					{
						Name:               "test-broker",
						Namespace:          "test-ns",
						UID:                "test-broker-uid",
						Labels:             map[string]string{"test-broker-label": "foo"},
						Annotations:        map[string]string{"test-broker-annotation": "foo"},
						ProvidedEventTypes: []string{"test-ns/test-eventtype-type"}},
				},
				EventTypes: []*EventType{
					{
						Name:        "test-eventtype",
						Namespace:   "test-ns",
						Type:        "test-eventtype-type",
						UID:         "test-eventtype-uid",
						Description: "test-eventtype-description",
						SchemaData:  "test-eventtype-schema-data",
						SchemaURL:   "http://test-eventtype-schema",
						Labels:      map[string]string{"test-eventtype-label": "foo"},
						Annotations: map[string]string{"test-eventtype-annotation": "foo"},
					},
				},
			},
			error: false,
		},
	}
	for _, tt := range tests {
		logger := zap.NewNop().Sugar()
		v1beta2objects := make([]runtime.Object, 0, 10)
		for _, et := range tt.eventTypes {
			v1beta2objects = append(v1beta2objects, et)
		}
		fakelistersv1beta2 := reconcilertestingv1beta2.NewListers(v1beta2objects)

		v1objects := make([]runtime.Object, 0, 10)
		for _, b := range tt.brokers {
			v1objects = append(v1objects, b)
		}
		fakelistersv1 := reconcilertestingv1.NewListers(v1objects)

		listers := Listers{
			BrokerLister:    fakelistersv1.GetBrokerLister(),
			EventTypeLister: fakelistersv1beta2.GetEventTypeLister(),
		}

		t.Run(tt.name, func(t *testing.T) {
			err, got := BuildEventMesh(listers, logger)
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

func brokerReference(brokerName, namespace string) *duckv1.KReference {
	return &duckv1.KReference{
		APIVersion: "eventing.knative.dev/v1",
		Kind:       "Broker",
		Name:       brokerName,
		Namespace:  namespace,
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
