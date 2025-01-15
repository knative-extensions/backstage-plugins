package eventmesh

import (
	"context"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	middleware "github.com/oapi-codegen/nethttp-middleware"

	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/injection"
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

	noTokenConfig := injection.ParseAndGetRESTConfigOrDie()

	noTokenConfig.BearerToken = ""
	noTokenConfig.Username = ""
	noTokenConfig.Password = ""
	noTokenConfig.BearerTokenFile = ""

	swagger, err := GetSwagger()
	if err != nil {
		log.Fatalf("Error loading swagger spec: %v", err)
	}

	endpoint := NewEndpoint(noTokenConfig, logger)
	strictHandler := NewStrictHandler(endpoint, []StrictMiddlewareFunc{})
	router := mux.NewRouter()
	router.Use(AuthTokenMiddleware())
	router.Use(requestValidator(swagger))
	handlerWithMiddleware := HandlerFromMux(strictHandler, router)

	r := kncloudevents.NewHTTPEventReceiver(8080)
	log.Fatal(r.StartListen(ctx, handlerWithMiddleware))
}

func requestValidator(swagger *openapi3.T) func(next http.Handler) http.Handler {
	return middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		Options: openapi3filter.Options{
			// we use a NoopAuthenticationFunc because we want to be able to
			// set the user in the context in the middleware
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	})
}
