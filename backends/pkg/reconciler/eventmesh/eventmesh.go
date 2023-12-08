package eventmesh

import (
	"context"

	pkgreconciler "knative.dev/pkg/reconciler"

	eventingv1beta2 "knative.dev/eventing/pkg/apis/eventing/v1beta2"
)

type Reconciler struct {
}

func (r *Reconciler) ReconcileKind(_ context.Context, _ *eventingv1beta2.EventType) pkgreconciler.Event {
	// TODO: do we actually need the reconciler?
	return nil
}
