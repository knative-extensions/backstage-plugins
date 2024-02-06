package eventmesh

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	"net/http"
	"sort"

	"knative.dev/pkg/injection/clients/dynamicclient"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventinglistersv1beta2 "knative.dev/eventing/pkg/client/listers/eventing/v1beta2"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/util/json"

	"knative.dev/pkg/logging"

	eventinglistersv1 "knative.dev/eventing/pkg/client/listers/eventing/v1"
)

// EventMesh is the top-level struct that holds the event mesh data.
// It's the struct that's serialized and sent to the Backstage plugin.
type EventMesh struct {
	// EventTypes is a list of all event types in the cluster.
	// While we can embed the event types in the brokers, we keep them separate because
	// not every event type is tied to a broker.
	EventTypes []*EventType `json:"eventTypes"`

	// Brokers is a list of all brokers in the cluster.
	Brokers []*Broker `json:"brokers"`
}

// Subscription is a set of Backstage IDs that are subscribed to a specific event type.
type Subscription struct {
	BackstageIds map[string]struct{}
}

// SubscriptionMap is a map of subscriptions that's created interim to build the EventMesh.
// key: "<namespace>/<eventType.spec.type>"
type SubscriptionMap map[string]Subscription

// BackstageLabel is the label that's used to identify Backstage resources.
// In Backstage Kubernetes plugin, a Backstage entity (e.g. a service) is tied to a Kubernetes resource
// using this label.
// see Backstage Kubernetes plugin for more details.
const BackstageLabel = "backstage.io/kubernetes-id"

// HttpHandler is the HTTP handler that's used to serve the event mesh data.
func HttpHandler(ctx context.Context, listers Listers) func(w http.ResponseWriter, req *http.Request) {
	logger := logging.FromContext(ctx)

	// this handler simply calls the event mesh builder and returns the result as JSON
	return func(w http.ResponseWriter, req *http.Request) {
		logger.Debugw("Handling request", "method", req.Method, "url", req.URL)

		eventMesh, err := BuildEventMesh(ctx, listers, logger)
		if err != nil {
			logger.Errorw("Error building event mesh", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(eventMesh)
		if err != nil {
			logger.Errorw("Error encoding event mesh", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// BuildEventMesh builds the event mesh data by fetching and converting the Kubernetes resources.
// The procedure is as follows:
// - Fetch the brokers and convert them to the representation that's consumed by the Backstage plugin.
// - Do the same for event types.
// - Fetch the triggers, find out what event types they're subscribed to and find out the resources that are receiving the events.
// - Make a connection between the event types and the subscribers. Store this connection in the eventType struct.
func BuildEventMesh(ctx context.Context, listers Listers, logger *zap.SugaredLogger) (EventMesh, error) {
	// fetch the brokers and convert them to the representation that's consumed by the Backstage plugin.
	convertedBrokers, err := fetchBrokers(listers.BrokerLister, logger)
	if err != nil {
		logger.Errorw("Error fetching and converting brokers", "error", err)
		return EventMesh{}, err
	}

	// build a map for easier access.
	// we need this map to register the event types in the brokers when we are processing the event types.
	// map key: "<namespace>/<name>"
	brokerMap := make(map[string]*Broker)
	for _, cbr := range convertedBrokers {
		brokerMap[cbr.GetNameAndNamespace()] = cbr
	}

	// fetch the event types and convert them to the representation that's consumed by the Backstage plugin.
	convertedEventTypes, err := fetchEventTypes(listers.EventTypeLister, logger)
	if err != nil {
		logger.Errorw("Error fetching and converting event types", "error", err)
		return EventMesh{}, err
	}

	// register the event types in the brokers
	for _, et := range convertedEventTypes {
		if et.Reference != "" {
			if br, ok := brokerMap[et.Reference]; ok {
				br.ProvidedEventTypes = append(br.ProvidedEventTypes, et.NameAndNamespace())
			}
		}
	}

	// build 2 maps for event types for easier access

	// map key: "<namespace>/<eventType.spec.type>"
	etTypeMap := make(map[string]*EventType)
	// map key: "<namespace>/<eventType.name>"
	etByNamespacedName := make(map[string]*EventType)

	for _, et := range convertedEventTypes {
		etTypeMap[et.NamespaceAndType()] = et
		etByNamespacedName[et.NameAndNamespace()] = et
	}

	subscriptionMap, err := buildSubscriptions(ctx, listers.TriggerLister, brokerMap, etByNamespacedName, logger)
	if err != nil {
		logger.Errorw("Error building subscriptions", "error", err)
		return EventMesh{}, err
	}

	for key, sub := range *subscriptionMap {
		for backstageId := range sub.BackstageIds {
			// find the event type and add the subscriber to the ConsumedBy list
			if et, ok := etTypeMap[key]; ok {
				et.ConsumedBy = append(et.ConsumedBy, backstageId)
			}
		}
	}

	eventMesh := EventMesh{
		EventTypes: convertedEventTypes,
		Brokers:    convertedBrokers,
	}

	return eventMesh, nil
}

func fetchBrokers(brokerLister eventinglistersv1.BrokerLister, logger *zap.SugaredLogger) ([]*Broker, error) {
	fetchedBrokers, err := brokerLister.List(labels.Everything())
	if err != nil {
		logger.Errorw("Error listing brokers", "error", err)
		return nil, err
	}

	convertedBrokers := make([]*Broker, 0, len(fetchedBrokers))
	for _, br := range fetchedBrokers {
		convertedBroker := convertBroker(br)
		convertedBrokers = append(convertedBrokers, &convertedBroker)
	}
	return convertedBrokers, err
}

func fetchEventTypes(eventTypeLister eventinglistersv1beta2.EventTypeLister, logger *zap.SugaredLogger) ([]*EventType, error) {
	fetchedEventTypes, err := eventTypeLister.List(labels.Everything())
	if err != nil {
		logger.Errorw("Error listing eventTypes", "error", err)
		return nil, err
	}

	sort.Slice(fetchedEventTypes, func(i, j int) bool {
		if fetchedEventTypes[i].Namespace != fetchedEventTypes[j].Namespace {
			return fetchedEventTypes[i].Namespace < fetchedEventTypes[j].Namespace
		}
		return fetchedEventTypes[i].Name < fetchedEventTypes[j].Name
	})

	convertedEventTypes := make([]*EventType, 0, len(fetchedEventTypes))
	for _, et := range fetchedEventTypes {
		convertedEventType := convertEventType(et)
		convertedEventTypes = append(convertedEventTypes, &convertedEventType)
	}

	return convertedEventTypes, err
}

func buildSubscriptions(ctx context.Context, triggerLister eventinglistersv1.TriggerLister, brokerMap map[string]*Broker, etByNamespacedName map[string]*EventType, logger *zap.SugaredLogger) (*SubscriptionMap, error) {
	// map key: "<namespace>/<eventType.spec.type>"
	subscriptionMap := make(SubscriptionMap)

	dynamicClient := dynamicclient.Get(ctx)

	triggers, err := triggerLister.List(labels.Everything())
	if err != nil {
		logger.Errorw("Error listing triggers", "error", err)
		return &subscriptionMap, err
	}

	for _, trigger := range triggers {
		// if the trigger's broker is not set or if we haven't processed the broker, we can skip the trigger
		if trigger.Spec.Broker == "" {
			continue
		}
		brokerRef := NameAndNamespace(trigger.Namespace, trigger.Spec.Broker)
		if _, ok := brokerMap[brokerRef]; !ok {
			continue
		}

		// if the trigger has no subscriber, we can skip it, there's no relation to show on Backstage side
		if trigger.Spec.Subscriber.Ref == nil {
			continue
		}

		subscriberBackstageId, err := getSubscriberBackstageId(ctx, dynamicClient, trigger, logger)
		if err != nil {
			// do not stop the Backstage plugin from rendering the rest of the data, e.g. because
			// there are no permissions to get a single subscriber resource
			continue
		}

		// we only care about subscribers that are in Backstage
		if len(subscriberBackstageId) == 0 {
			continue
		}

		// build the list of event types that the subscriber is subscribed to
		subscribedEventTypes := buildSubscribedEventTypes(trigger, brokerMap[brokerRef], etByNamespacedName, logger)

		// go over the event types and add the subscriber to the subscription map
		for _, eventType := range subscribedEventTypes {
			key := NameAndNamespace(trigger.Namespace, eventType)
			if _, ok := subscriptionMap[key]; !ok {
				subscriptionMap[key] = Subscription{
					BackstageIds: make(map[string]struct{}),
				}
			}
			subscriptionMap[key].BackstageIds[subscriberBackstageId] = struct{}{}
		}

	}

	return &subscriptionMap, nil
}

func buildSubscribedEventTypes(trigger *eventingv1.Trigger, broker *Broker, etNameMap map[string]*EventType, logger *zap.SugaredLogger) []string {
	// TODO: we don't handle the CESQL yet
	if trigger.Spec.Filter != nil && len(trigger.Spec.Filter.Attributes) > 0 {
		// check if "type" attribute is present
		if subscribedEventType, ok := trigger.Spec.Filter.Attributes["type"]; ok {
			// it can be present but empty
			// in that case, we assume the trigger is subscribed to all event types
			if subscribedEventType != eventingv1.TriggerAnyFilter {
				// if type is present and not empty, that means the trigger is subscribed to a specific event type
				return []string{subscribedEventType}
			}
		}
	}

	// if no filter or type is specified, we assume the resource is interested in all event types that the broker provides
	subscribedEventTypes := make([]string, 0, len(broker.ProvidedEventTypes))
	for _, eventType := range broker.ProvidedEventTypes {
		if et, ok := etNameMap[eventType]; ok {
			subscribedEventTypes = append(subscribedEventTypes, et.Type)
		}
	}

	return subscribedEventTypes
}

func getSubscriberBackstageId(ctx context.Context, client dynamic.Interface, trigger *eventingv1.Trigger, logger *zap.SugaredLogger) (string, error) {
	refGvr, _ := meta.UnsafeGuessKindToResource(schema.GroupVersionKind{
		trigger.Spec.Subscriber.Ref.Group,
		trigger.Spec.Subscriber.Ref.APIVersion,
		trigger.Spec.Subscriber.Ref.Kind,
	})

	resource, err := client.Resource(refGvr).Namespace(trigger.Spec.Subscriber.Ref.Namespace).Get(ctx, trigger.Spec.Subscriber.Ref.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		logger.Debugw("Subscriber resource not found", "resource", trigger.Spec.Subscriber.Ref.Name)
		return "", nil
	}
	if err != nil {
		logger.Errorw("Error fetching resource", "error", err)
		return "", err
	}

	// check if the resource has the Backstage label
	return resource.GetLabels()[BackstageLabel], nil
}
