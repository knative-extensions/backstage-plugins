package eventmesh

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"knative.dev/eventing/pkg/client/clientset/versioned"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/logging"

	"knative.dev/pkg/injection/clients/dynamicclient"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"

	"k8s.io/apimachinery/pkg/util/json"
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

// BackstageKubernetesIDLabel is the label that's used to identify Backstage resources.
// In Backstage Kubernetes plugin, a Backstage entity (e.g. a service) is tied to a Kubernetes resource
// using this label.
// see Backstage Kubernetes plugin for more details.
const BackstageKubernetesIDLabel = "backstage.io/kubernetes-id"

// HttpHandler is the HTTP handler that's used to serve the event mesh data.
type HttpHandler struct {
	ctx             context.Context
	inClusterConfig *rest.Config
}

// This handler simply calls the event mesh builder and returns the result as JSON
func (h HttpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := logging.FromContext(h.ctx)

	w.Header().Add("Content-Type", "application/json")

	logger.Debugw("Handling request", "method", req.Method, "url", req.URL)

	config := rest.CopyConfig(h.inClusterConfig)
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}
	// header value is in this format: "Bearer <token>"
	// we only need the token part
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		http.Error(w, "Invalid Authorization header. Should start with `Bearer `", http.StatusUnauthorized)
		return
	}
	config.BearerToken = authHeader[7:]
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	eventMesh, err := BuildEventMesh(h.ctx, clientset, logger)
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

// BuildEventMesh builds the event mesh data by fetching and converting the Kubernetes resources.
// The procedure is as follows:
// - Fetch the brokers and convert them to the representation that's consumed by the Backstage plugin.
// - Do the same for event types.
// - Fetch the triggers, find out what event types they're subscribed to and find out the resources that are receiving the events.
// - Make a connection between the event types and the subscribers. Store this connection in the eventType struct.
func BuildEventMesh(ctx context.Context, clientset versioned.Interface, logger *zap.SugaredLogger) (EventMesh, error) {
	// fetch the brokers and convert them to the representation that's consumed by the Backstage plugin.
	convertedBrokers, err := fetchBrokers(clientset, logger)
	if err != nil {
		logger.Errorw("Error fetching and converting brokers", "error", err)
		return EventMesh{}, err
	}

	// build a map for easier access.
	// we need this map to register the event types in the brokers when we are processing the event types.
	// map key: "<namespace>/<name>"
	brokerMap := make(map[string]*Broker)
	for _, cbr := range convertedBrokers {
		brokerMap[cbr.GetNamespacedName()] = cbr
	}

	// fetch the event types and convert them to the representation that's consumed by the Backstage plugin.
	convertedEventTypes, err := fetchEventTypes(clientset, logger)
	if err != nil {
		logger.Errorw("Error fetching and converting event types", "error", err)
		return EventMesh{}, err
	}

	// register the event types in the brokers
	for _, et := range convertedEventTypes {
		if et.Reference != "" {
			if br, ok := brokerMap[et.Reference]; ok {
				br.ProvidedEventTypes = append(br.ProvidedEventTypes, et.NamespacedName())
			}
		}
	}

	// fetch the triggers we will process them later
	triggers, err := clientset.EventingV1().Triggers(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})

	if err != nil {
		logger.Errorw("Error listing triggers", "error", err)
		return EventMesh{}, err
	}

	// build a map for easier access to the ETs by their namespaced name.
	// we need this map when processing the triggers to find out ET definitions for the ET references
	// brokers provide.
	// map key: "<namespace>/<eventType.name>"
	etByNamespacedName := make(map[string]*EventType)
	for _, et := range convertedEventTypes {
		etByNamespacedName[et.NamespacedName()] = et
	}

	for _, trigger := range triggers.Items {
		err := processTrigger(ctx, &trigger, brokerMap, etByNamespacedName, logger)
		if err != nil {
			logger.Errorw("Error processing trigger", "error", err)
			// do not stop the Backstage plugin from rendering the rest of the data, e.g. because
			// there are no permissions to get a single subscriber resource
		}
	}

	eventMesh := EventMesh{
		EventTypes: convertedEventTypes,
		Brokers:    convertedBrokers,
	}

	return eventMesh, nil
}

// processTrigger processes the trigger and updates the ETs that the trigger is subscribed to.
// The consumedBy fields of ETs are updated with the subscriber's Backstage ID.
func processTrigger(ctx context.Context, trigger *eventingv1.Trigger, brokerMap map[string]*Broker, etByNamespacedName map[string]*EventType, logger *zap.SugaredLogger) error {
	// if the trigger has no subscriber, we can skip it, there's no relation to show on Backstage side
	if trigger.Spec.Subscriber.Ref == nil {
		logger.Debugw("Trigger has no subscriber ref; cannot process this trigger", "namespace", trigger.Namespace, "trigger", trigger.Name)
		return nil
	}

	dynamicClient := dynamicclient.Get(ctx)
	subscriberBackstageId, err := getSubscriberBackstageId(ctx, dynamicClient, trigger.Spec.Subscriber.Ref, logger)
	if err != nil {
		// wrap the error to provide more context
		return fmt.Errorf("error getting subscriber backstage id: %w", err)
	}

	// we only care about subscribers that are in Backstage
	if len(subscriberBackstageId) == 0 {
		logger.Debugw("Subscriber has no backstage id", "namespace", trigger.Namespace, "trigger", trigger.Name)
		return nil
	}

	// if the trigger's broker is not set or if we haven't processed the broker, we can skip the trigger
	if trigger.Spec.Broker == "" {
		logger.Errorw("Trigger has no broker", "namespace", trigger.Namespace, "trigger", trigger.Name)
		return nil
	}
	brokerRef := NamespacedName(trigger.Namespace, trigger.Spec.Broker)
	if _, ok := brokerMap[brokerRef]; !ok {
		logger.Infow("Broker not found", "namespace", trigger.Namespace, "trigger", trigger.Name, "broker", trigger.Spec.Broker)
		return nil
	}

	eventTypes := collectSubscribedEventTypes(trigger, brokerMap[brokerRef], etByNamespacedName, logger)
	logger.Debugw("Collected subscribed event types", "namespace", trigger.Namespace, "trigger", trigger.Name, "broker", trigger.Spec.Broker, "eventTypes", eventTypes)

	for _, eventType := range eventTypes {
		eventType.ConsumedBy = append(eventType.ConsumedBy, subscriberBackstageId)
	}

	return nil
}

// collectSubscribedEventTypes collects the event types that the trigger is subscribed to.
// It does it by checking the trigger's filter and finding out the ET types that the filter is interested in.
// Later on, it finds the ETs that the broker provides and returns the ones matches the type.
// If the trigger has no filter, it returns all the ETs that the broker provides.
func collectSubscribedEventTypes(trigger *eventingv1.Trigger, broker *Broker, etByNamespacedName map[string]*EventType, logger *zap.SugaredLogger) []*EventType {
	logger.Debugw("Collecting subscribed event types", "namespace", trigger.Namespace, "trigger", trigger.Name, "broker", broker.Name)

	// TODO: we don't handle the CESQL yet
	if trigger.Spec.Filter != nil && len(trigger.Spec.Filter.Attributes) > 0 {
		logger.Debugw("Trigger has filter", "namespace", trigger.Namespace, "trigger", trigger.Name, "broker", broker.Name, "filter", trigger.Spec.Filter.Attributes)

		// check if "type" attribute is present
		if subscribedEventType, ok := trigger.Spec.Filter.Attributes["type"]; ok {
			logger.Debugw("Trigger has type filter", "namespace", trigger.Namespace, "trigger", trigger.Name, "broker", broker.Name, "type", subscribedEventType)

			// it can be present but empty
			// in that case, we assume the trigger is subscribed to all event types
			if subscribedEventType != eventingv1.TriggerAnyFilter {
				logger.Debugw("Trigger has non-empty type filter", "namespace", trigger.Namespace, "trigger", trigger.Name, "broker", broker.Name, "type", subscribedEventType)

				// if type is present and not empty, that means the trigger is subscribed to a ETs of that type
				// find the ETs for that type
				subscribedEventTypes := make([]*EventType, 0)
				for _, etNamespacedName := range broker.ProvidedEventTypes {
					if et, ok := etByNamespacedName[etNamespacedName]; ok {
						if et.Type == subscribedEventType {
							subscribedEventTypes = append(subscribedEventTypes, et)
						}
					}
				}
				logger.Debugw("Found subscribed event types", "namespace", trigger.Namespace, "trigger", trigger.Name, "broker", broker.Name, "subscribedEventTypes", subscribedEventTypes)
				return subscribedEventTypes
			}
		}
	}

	logger.Debugw("Trigger has no filter or type, returning all event types the broker provides", "namespace", trigger.Namespace, "trigger", trigger.Name, "broker", broker.Name)
	// if no filter or type is specified, we assume the resource is interested in all event types that the broker provides
	subscribedEventTypes := make([]*EventType, 0, len(broker.ProvidedEventTypes))
	for _, eventType := range broker.ProvidedEventTypes {
		if et, ok := etByNamespacedName[eventType]; ok {
			subscribedEventTypes = append(subscribedEventTypes, et)
		}
	}

	logger.Debugw("Found event types", "namespace", trigger.Namespace, "trigger", trigger.Name, "broker", broker.Name, "eventTypes", subscribedEventTypes)
	return subscribedEventTypes
}

// fetchBrokers fetches the brokers and converts them to the representation that's consumed by the Backstage plugin.
func fetchBrokers(clientset versioned.Interface, logger *zap.SugaredLogger) ([]*Broker, error) {
	brokers, err := clientset.EventingV1().Brokers(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})

	if err != nil {
		logger.Errorw("Error listing brokers", "error", err)
		return nil, err
	}

	convertedBrokers := make([]*Broker, 0, len(brokers.Items))
	for _, br := range brokers.Items {
		convertedBroker := convertBroker(&br)
		convertedBrokers = append(convertedBrokers, &convertedBroker)
	}
	return convertedBrokers, err
}

// fetchEventTypes fetches the event types and converts them to the representation that's consumed by the Backstage plugin.
func fetchEventTypes(clientset versioned.Interface, logger *zap.SugaredLogger) ([]*EventType, error) {
	eventTypeResponse, err := clientset.EventingV1beta2().EventTypes(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logger.Errorw("Error listing eventTypes", "error", err)
		return nil, err
	}
	eventTypes := eventTypeResponse.Items

	sort.Slice(eventTypes, func(i, j int) bool {
		if eventTypes[i].Namespace != eventTypes[j].Namespace {
			return eventTypes[i].Namespace < eventTypes[j].Namespace
		}
		return eventTypes[i].Name < eventTypes[j].Name
	})

	convertedEventTypes := make([]*EventType, 0, len(eventTypes))
	for _, et := range eventTypes {
		convertedEventType := convertEventType(&et)
		convertedEventTypes = append(convertedEventTypes, &convertedEventType)
	}

	return convertedEventTypes, err
}

// getSubscriberBackstageId fetches the subscriber resource and returns the Backstage ID if it's present.
func getSubscriberBackstageId(ctx context.Context, client dynamic.Interface, subRef *duckv1.KReference, logger *zap.SugaredLogger) (string, error) {
	refGvr, _ := meta.UnsafeGuessKindToResource(schema.FromAPIVersionAndKind(subRef.APIVersion, subRef.Kind))

	resource, err := client.Resource(refGvr).Namespace(subRef.Namespace).Get(ctx, subRef.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		logger.Debugw("Subscriber resource not found", "resource", subRef.Name)
		return "", nil
	}
	if err != nil {
		logger.Errorw("Error fetching resource", "error", err)
		return "", err
	}

	// check if the resource has the Backstage label
	if backstageId, ok := resource.GetLabels()[BackstageKubernetesIDLabel]; ok {
		return backstageId, nil
	}
	return "", nil
}
