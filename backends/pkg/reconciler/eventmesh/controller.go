package eventmesh

import (
	"context"
	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"
	"log"
)

func NewController(ctx context.Context) {

	logger := logging.FromContext(ctx)

	logger.Infow("Starting eventmesh-backend controller")

	startWebServer(ctx)
}

func startWebServer(ctx context.Context) {

	logger := logging.FromContext(ctx)

	logger.Infow("Starting eventmesh-backend webserver")

	noTokenConfig := injection.ParseAndGetRESTConfigOrDie()

	noTokenConfig.BearerToken = ""
	noTokenConfig.Username = ""
	noTokenConfig.Password = ""
	noTokenConfig.BearerTokenFile = ""

	r := kncloudevents.NewHTTPEventReceiver(8080)
	err := r.StartListen(ctx, HttpHandler{ctx, noTokenConfig})
	log.Fatal(err)
}
