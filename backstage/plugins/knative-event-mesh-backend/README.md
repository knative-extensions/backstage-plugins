# Knative Event Mesh plugin

The Event Mesh plugin is a Backstage plugin that allows you to view and manage Knative Eventing resources.

The Backstage plugin talks to a special backend that runs in the Kubernetes cluster and communicates with the Kubernetes
API server.

A demo setup for this plugin is available at https://github.com/aliok/knative-backstage-demo.

## Dynamic vs static plugin

This plugin has 2 distributions: static and dynamic.

The static distribution is a regular Backstage plugin that requires
the source code of Backstage to be changed.

The dynamic distribution is a plugin that can be installed without changing
the source code of Backstage. If you would like to use the dynamic plugin, please see the instructions in the
[Dynamic Plugin README file](./dist-dynamic/README.md).

Rest of this documentation is for the static plugin.

## Installation

Install the backend and the relevant configuration in the Kubernetes cluster

```bash
VERSION="latest" # or a specific version like 0.1.2
kubectl apply -f https://github.com/knative-extensions/backstage-plugins/releases/${VERSION}/download/eventmesh.yaml
```

In your Backstage directory, run the following command to install the plugin:

```bash
VERSION="latest" # or a specific version like 0.1.2
yarn workspace backend add @knative-extensions/plugin-knative-event-mesh-backend@${VERSION}
```

## Configuration

> **NOTE**: The backend needs to be accessible from the Backstage instance. If you are running the backend without
> exposing it, you can use `kubectl port-forward` to forward the port of the backend service to your local machine.
> ```bash
> kubectl port-forward -n knative-eventing svc/eventmesh-backend 8080:8080
> ```


The plugin needs to be configured to talk to the backend. It can be configured in the `app-config.yaml` file of the
Backstage instance and allows configuration of one or multiple providers.

Use a `knativeEventMesh` marker to start configuring the `app-config.yaml` file of Backstage:

```yaml
catalog:
  providers:
    knativeEventMesh:
      dev:
        token: '${KNATIVE_EVENT_MESH_TOKEN}'     # SA token to authenticate to the backend
        baseUrl: '${KNATIVE_EVENT_MESH_BACKEND}' # URL of the backend installed in the cluster
        schedule: # optional; same options as in TaskScheduleDefinition
          # supports cron, ISO duration, "human duration" as used in code
          frequency: { minutes: 1 }
          # supports ISO duration, "human duration" as used in code
          timeout: { minutes: 1 }
```

You can either manually change the placeholders in the `app-config.yaml` file or use environment variables to set the
values. The environment variables can be set as following before starting the Backstage instance:

```bash
export KNATIVE_EVENT_MESH_TOKEN=your-token
export KNATIVE_EVENT_MESH_BACKEND=http://localhost:8080
```

The value of `KNATIVE_EVENT_MESH_BACKEND` should be the URL of the backend service. If you are running the backend
service in the same cluster as the Backstage instance, you can use the service name as the URL. Or, if you are running
the backend without exposing it, you can use `kubectl port-forward` as mentioned above.

The value of `KNATIVE_EVENT_MESH_TOKEN` should be a service account token that has the necessary permissions to list
the Knative Eventing resources in the cluster. The backend will use this token to authenticate to the Kubernetes API
server. This is required for security reasons as otherwise (if the backend is running with a SA token directly) the
backend would have full access to the cluster will be returning all resources to anyone who can access the backend.

The token will require the following permissions to work properly:

- `get`, `list` and `watch` permissions for `eventing.knative.dev/brokers`, `eventing.knative.dev/eventtypes` and
  `eventing.knative.dev/triggers` resources
- `get` permission for all resources to fetch subscribers for triggers

You can create a ClusterRole with the necessary permissions and bind it to the service account token.

An example configuration is as follows:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-eventmesh-backend-service-account
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: my-eventmesh-backend-cluster-role
rules:
  # permissions for eventtypes, brokers and triggers
  - apiGroups:
      - "eventing.knative.dev"
    resources:
      - brokers
      - eventtypes
      - triggers
    verbs:
      - get
      - list
      - watch
  # permissions to get subscribers for triggers
  # as subscribers can be any resource, we need to give access to all resources
  # we fetch subscribers one by one, we only need `get` verb
  - apiGroups:
      - "*"
    resources:
      - "*"
    verbs:
      - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: my-eventmesh-backend-cluster-role-binding
subjects:
  - kind: ServiceAccount
    name: my-eventmesh-backend-service-account
    namespace: default
roleRef:
  kind: ClusterRole
  name: my-eventmesh-backend-cluster-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: Secret
metadata:
  name: my-eventmesh-backend-secret
  namespace: default
  annotations:
    kubernetes.io/service-account.name: my-eventmesh-backend-service-account
type: kubernetes.io/service-account-token
```

To get the token, you can run the following command:

```bash
kubectl get secret my-eventmesh-backend-secret -o jsonpath='{.data.token}' | base64 --decode
```

Run a quick check to see if the token works:

```bash
export KUBE_API_SERVER_URL=$(kubectl config view --minify --output jsonpath="{.clusters[*].cluster.server}") # e.g. "https://192.168.2.151:16443"
export KUBE_SA_TOKEN=$(kubectl get secret my-eventmesh-backend-secret -o jsonpath='{.data.token}' | base64 --decode)
curl -k -H "Authorization: Bearer $KUBE_SA_TOKEN" -X GET "${KUBE_API_SERVER_URL}/apis/eventing.knative.dev/v1/namespaces/default/brokers" | json_pp
# Should see the brokers, or nothing if there are no brokers
# But, should not see an error
```

Run a second sanity check to see if the token works with the backend

```bash
KNATIVE_EVENT_MESH_BACKEND=http://localhost:8080 # or the URL of the backend
export KUBE_SA_TOKEN=$(kubectl get secret my-eventmesh-backend-secret -o jsonpath='{.data.token}' | base64 --decode)
curl -k -H "Authorization: Bearer $KUBE_SA_TOKEN" -X GET "${KNATIVE_EVENT_MESH_BACKEND}" | json_pp
# Should see the response from the backend such as 
# {
#   "brokers" : [...],
#   "eventTypes" : [...]
#}
```

If these sanity checks work, you can use the token in the `app-config.yaml` file as the value
of `KNATIVE_EVENT_MESH_TOKEN`.

### Legacy Backend Installation

Configure the scheduler for the entity provider and enable the processor. Add the following code
to `packages/backend/src/plugins/catalog.ts` file:

```ts
import {CatalogClient} from "@backstage/catalog-client";
import {
    KnativeEventMeshProcessor,
    KnativeEventMeshProvider
} from '@knative-extensions/plugin-knative-event-mesh-backend';

export default async function createPlugin(
    env:PluginEnvironment,
):Promise<Router> {
    const builder = await CatalogBuilder.create(env);

    /* ... other processors and/or providers ... */

    // ADD THESE
    builder.addEntityProvider(
        KnativeEventMeshProvider.fromConfig(env.config, {
            logger: env.logger,
            scheduler: env.scheduler,
        }),
    );
    const catalogApi = new CatalogClient({
        discoveryApi: env.discovery,
    });
    const knativeEventMeshProcessor = new KnativeEventMeshProcessor(catalogApi, env.logger);
    builder.addProcessor(knativeEventMeshProcessor);

    /* ... other processors and/or providers ... */

    const {processingEngine, router} = await builder.build();
    await processingEngine.start();
    return router;
}
```

### New Backend Installation

To install on the new backend system, add the following into the `packages/backend/index.ts` file:

```ts title=packages/backend/index.ts
import { createBackend } from '@backstage/backend-defaults';

const backend = const backend = createBackend();

// Other plugins/modules

backend.add(import('@knative-extensions/plugin-knative-event-mesh-backend/alpha'));

```

> **NOTE**: If you have made any changes to the schedule in the `app-config.yaml` file, then restart to apply the
> changes.

## Troubleshooting

When you start your Backstage application, you can see some log lines as follows:

```text
[1] 2024-01-04T09:38:08.707Z knative-event-mesh-backend info Found 1 knative event mesh provider configs with ids: dev type=plugin
```

## Usage

The plugin will register a few entities in the Backstage catalog.

Screenshots:

- ![Event Mesh plugin](./event-mesh-plugin-components-view.png)

- ![Event Mesh plugin](./event-mesh-plugin-apis-view.png)
