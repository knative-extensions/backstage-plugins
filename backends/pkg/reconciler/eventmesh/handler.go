package eventmesh

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventinglistersv1beta2 "knative.dev/eventing/pkg/client/listers/eventing/v1beta2"
	"net/http"
	"sort"
	"strings"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/util/json"

	"knative.dev/pkg/logging"

	eventinglistersv1 "knative.dev/eventing/pkg/client/listers/eventing/v1"

	dynamicclient "knative.dev/pkg/injection/clients/dynamicclient"
)

type EventMesh struct {
	// not every event type is tied to a broker. thus, we need to send event types as well.
	EventTypes []*EventType `json:"eventTypes"`
	Brokers    []*Broker    `json:"brokers"`
}

type Subscription struct {
	BackstageIds map[string]struct{}
}

// SubscriptionMap key: "<namespace>/<eventType.spec.type>"
type SubscriptionMap map[string]Subscription

const BackstageLabel = "backstage.io/kubernetes-id"

func EventMeshHandler(ctx context.Context, listers Listers) func(w http.ResponseWriter, req *http.Request) {
	logger := logging.FromContext(ctx)

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

func BuildEventMesh(ctx context.Context, listers Listers, logger *zap.SugaredLogger) (EventMesh, error) {
	convertedBrokers, err := fetchBrokers(listers.BrokerLister, logger)
	if err != nil {
		logger.Errorw("Error fetching and converting brokers", "error", err)
		return EventMesh{}, err
	}

	// map key: "<namespace>/<name>"
	brokerMap := make(map[string]*Broker)
	for _, cbr := range convertedBrokers {
		brokerMap[cbr.GetNameAndNamespace()] = cbr
	}

	convertedEventTypes, err := fetchEventTypes(listers.EventTypeLister, logger)
	if err != nil {
		logger.Errorw("Error fetching and converting event types", "error", err)
		return EventMesh{}, err
	}

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
	etNameMap := make(map[string]*EventType)

	for _, et := range convertedEventTypes {
		etTypeMap[et.NamespaceAndType()] = et
		etNameMap[et.NameAndNamespace()] = et
	}

	subscriptionMap, err := buildSubscriptions(ctx, listers.TriggerLister, brokerMap, etNameMap, logger)
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

func buildSubscriptions(ctx context.Context, triggerLister eventinglistersv1.TriggerLister, brokerMap map[string]*Broker, etNameMap map[string]*EventType, logger *zap.SugaredLogger) (*SubscriptionMap, error) {
	// map key: "<namespace>/<eventType.spec.type>"
	subscriptionMap := make(SubscriptionMap)

	triggers, err := triggerLister.List(labels.Everything())
	if err != nil {
		logger.Errorw("Error listing triggers", "error", err)
		return nil, err
	}

	for _, trigger := range triggers {
		// if the trigger's broker is not set or if we haven't processed the broker, we can skip the trigger
		if trigger.Spec.Broker == "" {
			continue
		}
		brokerRef := NameAndNamespace(trigger.Namespace, trigger.Spec.Broker)
		if _, ok := brokerMap[brokerRef]; !ok {
			return nil, nil
		}

		// if the trigger has no subscriber, we can skip it, there's no relation to show on Backstage side
		if trigger.Spec.Subscriber.Ref == nil {
			return nil, nil
		}

		subscriberBackstageId, err := getSubscriberBackstageId(ctx, trigger, logger)
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
		subscribedEventTypes := buildSubscribedEventTypes(trigger, brokerMap[brokerRef], etNameMap, logger)

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
			// if type is present, that means the trigger is subscribed to a specific event type
			// get that event type
			return []string{subscribedEventType}
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

func getSubscriberBackstageId(ctx context.Context, trigger *eventingv1.Trigger, logger *zap.SugaredLogger) (string, error) {
	client := dynamicclient.Get(ctx)

	refGvr := schema.GroupVersionResource{
		Group:   trigger.Spec.Subscriber.Ref.Group,
		Version: trigger.Spec.Subscriber.Ref.APIVersion,
		// TODO: couldn't remember the elegant way to do this
		Resource: strings.ToLower(trigger.Spec.Subscriber.Ref.Kind) + "s",
	}

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
