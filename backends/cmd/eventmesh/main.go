package main

import (
	"knative.dev/backstage-plugins/backends/pkg/reconciler/eventmesh"

	"knative.dev/pkg/signals"
)

func main() {
	ctx := signals.NewContext()
	eventmesh.NewController(ctx)
}
