package eventmesh

import (
	"context"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	middleware "github.com/oapi-codegen/nethttp-middleware"

	"knative.dev/backstage-plugins/backends/pkg/eventmesh/auth"
	eventmeshv1 "knative.dev/backstage-plugins/backends/pkg/eventmesh/v1"

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

	v1swagger, err := eventmeshv1.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading swagger spec: %v", err)
	}
	// the paths in OpenAPI spec are not prefixed with /v1
	// but, we want to serve them at /v1
	// this spec is used by the request validator middleware
	prefixSwaggerPaths(v1swagger, "/v1")

	v1endpoint := eventmeshv1.NewEndpoint(noTokenConfig, logger)
	v1strictHandler := eventmeshv1.NewStrictHandler(v1endpoint, []eventmeshv1.StrictMiddlewareFunc{})
	v1router := mux.NewRouter()
	v1router.Use(auth.AuthTokenMiddleware())
	v1router.Use(requestValidator(v1swagger))
	v1handlerWithMiddleware := eventmeshv1.HandlerFromMuxWithBaseURL(v1strictHandler, v1router, "/v1")

	parentRouter := mux.NewRouter()
	parentRouter.PathPrefix("/v1/").Handler(v1handlerWithMiddleware)

	r := kncloudevents.NewHTTPEventReceiver(8080)
	log.Fatal(r.StartListen(ctx, parentRouter))
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

func prefixSwaggerPaths(swagger *openapi3.T, prefix string) {
	// iterate over swagger.Paths.InMatchingOrder()
	for _, pathKey := range swagger.Paths.InMatchingOrder() {
		pathItem := swagger.Paths.Value(pathKey)
		swagger.Paths.Set(prefix+pathKey, pathItem)
		swagger.Paths.Delete(pathKey)
	}
}
