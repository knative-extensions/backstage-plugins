package eventmesh

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/util/json"

	"knative.dev/pkg/logging"
)

type EventMesh struct {
	// not every event type is tied to a broker. thus, we need to send event types as well.
	EventTypes []*EventType `json:"eventTypes"`
	Brokers    []*Broker    `json:"brokers"`
	// TODO: triggers
	// Triggers   []Trigger   `json:"triggers"`
}

func EventMeshHandler(ctx context.Context) func(w http.ResponseWriter, req *http.Request) {
	logger := logging.FromContext(ctx)

	return func(w http.ResponseWriter, req *http.Request) {
		// TODO: build the mesh here
		eventMesh := EventMesh{
			EventTypes: []*EventType{},
			Brokers:    []*Broker{},
		}

		err := json.NewEncoder(w).Encode(eventMesh)
		if err != nil {
			logger.Errorw("Error encoding event mesh", "error", err)
			return
		}
	}
}
