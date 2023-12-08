package eventmesh

import (
	"context"
	"net/http"
	"slices"
	"sort"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/util/json"

	"knative.dev/pkg/logging"

	eventingb1beta2 "knative.dev/eventing/pkg/apis/eventing/v1beta2"
	eventinglistersv1 "knative.dev/eventing/pkg/client/listers/eventing/v1"
)

type EventMesh struct {
	// not every event type is tied to a broker. thus, we need to send event types as well.
	EventTypes []*EventType `json:"eventTypes"`
	Brokers    []*Broker    `json:"brokers"`
	// TODO: triggers
	// Triggers   []Trigger   `json:"triggers"`
}

type EventTypeMap = map[string]*EventType

func EventMeshHandler(ctx context.Context, listers Listers) func(w http.ResponseWriter, req *http.Request) {
	logger := logging.FromContext(ctx)

	return func(w http.ResponseWriter, req *http.Request) {
		logger.Debugw("Handling request", "method", req.Method, "url", req.URL)

		err, eventMesh := BuildEventMesh(listers, logger)
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

func BuildEventMesh(listers Listers, logger *zap.SugaredLogger) (error, EventMesh) {
	err, convertedBrokers := fetchBrokers(listers.BrokerLister, logger)
	if err != nil {
		logger.Errorw("Error fetching and converting brokers", "error", err)
		return err, EventMesh{}
	}

	brokerMap := make(map[string]*Broker)
	for _, cbr := range convertedBrokers {
		brokerMap[cbr.GetNameAndNamespace()] = cbr
	}

	fetchedEventTypes, err := listers.EventTypeLister.List(labels.Everything())
	if err != nil {
		logger.Errorw("Error listing eventTypes", "error", err)
		return err, EventMesh{}
	}

	sort.Slice(fetchedEventTypes, func(i, j int) bool {
		if fetchedEventTypes[i].Namespace != fetchedEventTypes[j].Namespace {
			return fetchedEventTypes[i].Namespace < fetchedEventTypes[j].Namespace
		}
		return fetchedEventTypes[i].Name < fetchedEventTypes[j].Name
	})

	logger.Debugw("Fetched event types", "event types", fetchedEventTypes)

	convertedEventTypeMap := make(EventTypeMap)
	for _, et := range fetchedEventTypes {
		namespaceEventTypeRef := NamespaceEventTypeRef(et)

		if et.Spec.Reference != nil {
			if br, ok := brokerMap[RefNameAndNamespace(et.Spec.Reference)]; ok {
				// add to broker provided event types
				// only add if it hasn't been added already
				if !slices.Contains(br.ProvidedEventTypes, namespaceEventTypeRef) {
					br.ProvidedEventTypes = append(br.ProvidedEventTypes, namespaceEventTypeRef)
				}
			}
		}

		if _, ok := convertedEventTypeMap[namespaceEventTypeRef]; ok {
			logger.Debugw("Duplicate event type", "event type", namespaceEventTypeRef)
			continue
		}

		convertedEventType := convertEventType(et)
		convertedEventTypeMap[namespaceEventTypeRef] = &convertedEventType
	}

	eventMesh := EventMesh{
		EventTypes: make([]*EventType, 0, len(convertedEventTypeMap)),
		Brokers:    convertedBrokers,
	}

	for _, et := range convertedEventTypeMap {
		eventMesh.EventTypes = append(eventMesh.EventTypes, et)
	}
	return nil, eventMesh
}

func fetchBrokers(brokerLister eventinglistersv1.BrokerLister, logger *zap.SugaredLogger) (error, []*Broker) {
	fetchedBrokers, err := brokerLister.List(labels.Everything())
	if err != nil {
		logger.Errorw("Error listing brokers", "error", err)
		return err, nil
	}

	convertedBrokers := make([]*Broker, 0, len(fetchedBrokers))
	for _, br := range fetchedBrokers {
		convertedBroker := convertBroker(br)
		convertedBrokers = append(convertedBrokers, &convertedBroker)
	}
	return err, convertedBrokers
}

func NamespaceEventTypeRef(et *eventingb1beta2.EventType) string {
	return BuildNamespaceEventTypeRef(et.Namespace, et.Spec.Type)
}

func BuildNamespaceEventTypeRef(namespace, eventType string) string {
	return namespace + "/" + eventType
}
