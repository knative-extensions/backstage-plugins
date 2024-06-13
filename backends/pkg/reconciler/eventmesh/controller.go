package eventmesh

import (
	"context"
	"log"

	"k8s.io/client-go/rest"

	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/logging"
)

func NewController(ctx context.Context) {

	logger := logging.FromContext(ctx)

	logger.Infow("Starting eventmesh-backend controller")

	startWebServer(ctx)
}

func startWebServer(ctx context.Context) {

	logger := logging.FromContext(ctx)

	logger.Infow("Starting eventmesh-backend webserver")

	noTokenConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error getting in-cluster config: %v", err)
	}

	noTokenConfig.BearerToken = ""
	noTokenConfig.Username = ""
	noTokenConfig.Password = ""
	noTokenConfig.BearerTokenFile = ""

	r := kncloudevents.NewHTTPEventReceiver(8080)
	err = r.StartListen(ctx, HttpHandler{ctx, noTokenConfig})
	log.Fatal(err)
}
