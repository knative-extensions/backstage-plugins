package eventmesh

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	eventtypereconciler "knative.dev/eventing/pkg/client/injection/reconciler/eventing/v1beta2/eventtype"
)

func NewController(ctx context.Context) *controller.Impl {

	reconciler := &Reconciler{}

	logger := logging.FromContext(ctx)

	logger.Infow("Starting eventmesh-backend controller")

	// shared main does all the injection and starts the controller
	// thus, we want to use it.
	// and, it wants a controller.Impl, so, we're just returning one that's not really used in reality.
	impl := eventtypereconciler.NewImpl(ctx, reconciler)

	return impl
}

func startWebServer(ctx context.Context) {

	logger := logging.FromContext(ctx)

	logger.Infow("Starting eventmesh-backend webserver")

	r := mux.NewRouter()
	r.Use(commonMiddleware)

	r.HandleFunc("/", EventMeshHandler(ctx)).Methods("GET")
	http.Handle("/", r)

	// TODO: port
	log.Fatal(http.ListenAndServe(":8000", r))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
