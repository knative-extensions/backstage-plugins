package eventmesh

import (
	"context"

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
