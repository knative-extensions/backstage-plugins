package eventmesh

import (
	"context"

	pkgreconciler "knative.dev/pkg/reconciler"

	eventingv1beta2 "knative.dev/eventing/pkg/apis/eventing/v1beta2"
)

type Reconciler struct {
}

func (r *Reconciler) ReconcileKind(_ context.Context, _ *eventingv1beta2.EventType) pkgreconciler.Event {
	// we don't actually need a reconciler.
	// but the sharedmain from knative/pkg requires one to inject informers, which we use for the eventmesh.
	return nil
}
