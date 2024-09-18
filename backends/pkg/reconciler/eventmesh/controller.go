package eventmesh

import (
	"context"
	"log"
	"os"

	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"

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

	var noTokenConfig *rest.Config
	var err error

	// if KUBECONFIG env var is present, use it, instead of in-cluster config
	// this is especially useful for local development
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath != "" {
		noTokenConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %v", err)
		}
	} else {
		noTokenConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error getting in-cluster config: %v", err)
		}
	}

	noTokenConfig.BearerToken = ""
	noTokenConfig.Username = ""
	noTokenConfig.Password = ""
	noTokenConfig.BearerTokenFile = ""

	r := kncloudevents.NewHTTPEventReceiver(8080)
	err = r.StartListen(ctx, HttpHandler{ctx, noTokenConfig})
	log.Fatal(err)
}
