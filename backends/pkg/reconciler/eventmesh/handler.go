package eventmesh

import (
	"context"
	eventinglistersv1beta2 "knative.dev/eventing/pkg/client/listers/eventing/v1beta2"
	"net/http"
	"sort"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/util/json"

	"knative.dev/pkg/logging"

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

		eventMesh, err := BuildEventMesh(listers, logger)
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

func BuildEventMesh(listers Listers, logger *zap.SugaredLogger) (EventMesh, error) {
	convertedBrokers, err := fetchBrokers(listers.BrokerLister, logger)
	if err != nil {
		logger.Errorw("Error fetching and converting brokers", "error", err)
		return EventMesh{}, err
	}

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
