//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/knative"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/eventshub/assert"
	"knative.dev/reconciler-test/pkg/feature"
)

func TestIntegration(t *testing.T) {
	t.Parallel()

	ctx, env := global.Environment(
		knative.WithKnativeNamespace("knative-eventing"),
		knative.WithLoggingConfig,
		knative.WithTracingConfig,
		k8s.WithEventListener,
		environment.Managed(t),
	)
	env.Test(ctx, t, VerifyBackstageBackendAuthentication())
}

func VerifyBackstageBackendAuthentication() *feature.Feature {

	f := feature.NewFeature()

	authenticatedClientName := feature.MakeRandomK8sName("authenticated-client")
	unauthenticatedClientName := feature.MakeRandomK8sName("unauthenticated-client")
	SANamespace := "eventmesh-backend-user-namespace"
	SecretName := "eventmesh-backend-user-secret"

	f.Setup("request with authenticated client", func(ctx context.Context, t feature.T) {
		secret, err := kubeclient.Get(ctx).CoreV1().Secrets(SANamespace).Get(ctx, SecretName, metav1.GetOptions{})
		if err != nil {
			t.Fatal("Failed to get secret", err)
		}

		token := string(secret.Data["token"])

		eventshub.Install(authenticatedClientName,
			eventshub.StartSenderURL("http://eventmesh-backend.knative-eventing.svc.cluster.local:8080/v1/getEventMesh"),
			eventshub.InputHeader("Authorization", "Bearer "+token),
			eventshub.InputMethod("GET"),
		)(ctx, t)
	})
	f.Setup("request with unauthenticated client", eventshub.Install(
		unauthenticatedClientName,
		eventshub.StartSenderURL("http://eventmesh-backend.knative-eventing.svc.cluster.local:8080/v1/getEventMesh"),
		eventshub.InputHeader("Foo", "Bar"),
		eventshub.InputMethod("GET")),
	)

	f.Assert("assert response with authenticated client", assert.OnStore(authenticatedClientName).
		Match(assert.MatchKind(eventshub.EventResponse)).
		Match(assert.MatchStatusCode(200)).
		AtLeast(1))
	f.Assert("assert response with unauthenticated client", assert.OnStore(unauthenticatedClientName).
		Match(assert.MatchKind(eventshub.EventResponse)).
		Match(assert.MatchStatusCode(401)).
		AtLeast(1))

	return f
}
