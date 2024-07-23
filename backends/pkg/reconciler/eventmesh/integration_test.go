package eventmesh

import (
	"context"
	"testing"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/eventshub/assert"
	"knative.dev/reconciler-test/pkg/feature"
)

var global environment.GlobalEnvironment

func TestIntegration(t *testing.T) {
	t.Parallel()

	ctx, env := global.Environment(environment.Managed(t))
	env.Test(ctx, t, TesAuth())
}

func TesAuth() *feature.Feature {
	f := feature.NewFeature()

	authenticatedClientName := feature.MakeRandomK8sName("authenticated-client")
	unauthenticatedClientName := feature.MakeRandomK8sName("unauthenticated-client")

	SAName := "sa"
	SANamespace := "default"

	f.Setup("request with authenticated client", func(ctx context.Context, t feature.T) {
		tr := &authenticationv1.TokenRequest{
			// TODO: fill up
		}
		tr, err := kubeclient.Get(ctx).
			CoreV1().
			ServiceAccounts(SANamespace).
			CreateToken(ctx, SAName, tr, metav1.CreateOptions{})
		if err != nil {
			t.Fatal("Failed to create token for SA", err)
		}
		eventshub.Install(authenticatedClientName,
			eventshub.StartSenderURL("TODO_backstage_backend_url"),
			eventshub.InputHeader("Authorization", "Bearer "+tr.Status.Token),
			eventshub.InputMethod("GET"),
		)(ctx, t)
	})

	f.Setup("request with unauthenticated client", eventshub.Install(
		unauthenticatedClientName,
		eventshub.StartSenderURL("localhost:8080"),
		eventshub.InputMethod("GET")),
	)

	f.Assert("assert response with authenticated client", assert.OnStore(authenticatedClientName).
		Match(assert.MatchKind(eventshub.EventResponse)).
		Match(assert.MatchStatusCode(202)).
		AtLeast(1))
	f.Assert("assert response with unauthenticated client", assert.OnStore(unauthenticatedClientName).
		Match(assert.MatchKind(eventshub.EventResponse)).
		Match(assert.MatchStatusCode(401)).
		AtLeast(1))

	return f
}
