package eventmesh

import (
	"context"
	"log"
	"net/http"

	"k8s.io/client-go/rest"

	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/logging"
)

func NewController(ctx context.Context) {

	logger := logging.FromContext(ctx)

	logger.Infow("Starting eventmesh-backend controller")

	// TODO: does not stop with SIGTERM
	startWebServer(ctx)
}

func startWebServer(ctx context.Context) {

	logger := logging.FromContext(ctx)

	logger.Infow("Starting eventmesh-backend webserver")

	noTokenConfig, err := rest.InClusterConfig()
	noTokenConfig.BearerToken = ""
	noTokenConfig.Username = ""
	noTokenConfig.Password = ""
	noTokenConfig.BearerTokenFile = ""

	if err != nil {
		log.Fatalf("Error getting in-cluster config: %v", err)
	}
	r := kncloudevents.NewHTTPEventReceiver(8080)
	err = r.StartListen(ctx, HttpHandler{ctx, noTokenConfig})
	log.Fatal(err)

	// TODO: err

	//r.HandleFunc("/", HttpHandler(ctx, noTokenConfig)).Methods("GET")
	//http.Handle("/", r)

	//log.Fatal(http.ListenAndServe(":8080", r))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
