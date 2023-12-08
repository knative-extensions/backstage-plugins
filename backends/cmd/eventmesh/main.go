package main

import (
	"context"

	"knative.dev/backstage-plugins/backends/pkg/reconciler/eventmesh"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
)

const (
	component = "eventmesh-backend"
)

func main() {

	sharedmain.MainNamed(signals.NewContext(), component,

		injection.NamedControllerConstructor{
			Name: "backend",
			ControllerConstructor: func(ctx context.Context, watcher configmap.Watcher) *controller.Impl {
				return eventmesh.NewController(ctx)
			},
		},
	)
}
