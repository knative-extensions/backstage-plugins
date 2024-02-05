package eventmesh

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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
	EventTypes  []*EventType  `json:"eventTypes"`
	Brokers     []*Broker     `json:"brokers"`
	Subscribers []*Subscriber `json:"subscribers"`
}

const BackstageLabel = "backstage.io/kubernetes-id"

type EventTypeMap = map[string]*EventType

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

	subscriberMap, err := buildSubscriberMap(ctx, listers.TriggerLister, logger, brokerMap, convertedEventTypes)

	// convert subscriber map to slice
	subscribers := make([]*Subscriber, 0, len(*subscriberMap))
	for _, sub := range *subscriberMap {
		subscribers = append(subscribers, sub)
	}

	eventMesh := EventMesh{
		EventTypes:  convertedEventTypes,
		Brokers:     convertedBrokers,
		Subscribers: subscribers,
	}

	return eventMesh, nil
}

func fetchBrokers(brokerLister eventinglistersv1.BrokerLister, logger *zap.SugaredLogger) ([]*Broker, error) {
	fetchedBrokers, err := brokerLister.List(labels.Everything())
	// TODO handle 404?
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
	// TODO handle 404?
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

func buildSubscriberMap(ctx context.Context, triggerLister eventinglistersv1.TriggerLister, logger *zap.SugaredLogger, brokerMap map[string]*Broker, eventTypes []*EventType) (*map[string]*Subscriber, error) {
	// map of<UID, Subscriber>
	subscriberMap := make(map[string]*Subscriber)

	triggers, err := triggerLister.List(labels.Everything())
	if err != nil {
		logger.Errorw("Error listing triggers", "error", err)
		return nil, err
	}

	client := dynamicclient.Get(ctx)

	for _, trigger := range triggers {
		subscriber, err := processTrigger(ctx, trigger, brokerMap, eventTypes, client, logger)

		if err != nil {
			logger.Errorw("Error processing trigger", "error", err)
			// do not stop the Backstage plugin from rendering the rest of the data, e.g. because
			// there are no permissions to get a single subscriber resource
			continue
		}

		// if the subscriber is not in the map, we add it
		subscriberId := subscriber.BackstageId
		if _, ok := subscriberMap[subscriberId]; !ok {
			subscriberMap[subscriberId] = subscriber
		}

		// need to add the newly found subscribed event types to the subscriber
		subscriberMap[subscriberId].SubscribedEventTypes = append(subscriberMap[subscriberId].SubscribedEventTypes, subscriber.SubscribedEventTypes...)
	}

	// deduplicate the event types in the subscribers
	for _, sub := range subscriberMap {
		sub.SubscribedEventTypes = Deduplicate(sub.SubscribedEventTypes)
	}

	return &subscriberMap, nil
}

func processTrigger(ctx context.Context, trigger *eventingv1.Trigger, brokerMap map[string]*Broker, eventTypes []*EventType, client dynamic.Interface, logger *zap.SugaredLogger) (*Subscriber, error) {
	// if the trigger's broker is not set or if we haven't processed the broker, we can skip the trigger
	if trigger.Spec.Broker == "" {
		return nil, nil
	}

	brokerRef := NameAndNamespace(trigger.Namespace, trigger.Spec.Broker)
	if _, ok := brokerMap[brokerRef]; !ok {
		return nil, nil
	}

	// if the trigger has no subscriber, we can skip it, there's no relation to show on Backstage side
	if trigger.Spec.Subscriber.Ref == nil {
		return nil, nil
	}

	refGvr := schema.GroupVersionResource{
		Group:   trigger.Spec.Subscriber.Ref.Group,
		Version: trigger.Spec.Subscriber.Ref.APIVersion,
		// TODO: couldn't remember the elegant way to do this
		Resource: strings.ToLower(trigger.Spec.Subscriber.Ref.Kind) + "s",
	}

	resource, err := client.Resource(refGvr).Namespace(trigger.Spec.Subscriber.Ref.Namespace).Get(ctx, trigger.Spec.Subscriber.Ref.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		logger.Debugw("Subscriber resource not found", "resource", trigger.Spec.Subscriber.Ref.Name)
		return nil, nil
	}
	if err != nil {
		logger.Errorw("Error fetching resource", "error", err)
		return nil, err
	}

	// check if the resource has the Backstage label
	backstageId, ok := resource.GetLabels()[BackstageLabel]
	if !ok {
		return nil, nil
	}

	var subscribedEventTypes []string

	// TODO: we don't handle the CESQL yet
	if trigger.Spec.Filter != nil && len(trigger.Spec.Filter.Attributes) > 0 {
		// check if "type" attribute is present
		if subscribedEventType, ok := trigger.Spec.Filter.Attributes["type"]; ok {
			// iterate over found event types and find the ones that match this one
			for _, et := range eventTypes {
				if et.Type == subscribedEventType {
					subscribedEventTypes = append(subscribedEventTypes, et.NameAndNamespace())
				}
			}
		}
	}

	if len(subscribedEventTypes) == 0 {
		// if no type is specified, we assume the resource is interested in all event types that the broker provides
		subscribedEventTypes = brokerMap[brokerRef].ProvidedEventTypes
	}

	s := &Subscriber{
		Group:                trigger.Spec.Subscriber.Ref.Group,
		Version:              trigger.Spec.Subscriber.Ref.APIVersion,
		Kind:                 trigger.Spec.Subscriber.Ref.Kind,
		Namespace:            trigger.Spec.Subscriber.Ref.Namespace,
		Name:                 trigger.Spec.Subscriber.Ref.Name,
		UID:                  string(resource.GetUID()),
		Labels:               resource.GetLabels(),
		Annotations:          FilterAnnotations(resource.GetAnnotations()),
		SubscribedEventTypes: subscribedEventTypes,
		BackstageId:          backstageId,
	}
	return s, nil
}
