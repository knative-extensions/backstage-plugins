# Knative Event Mesh plugin

The Event Mesh plugin is a Backstage plugin that allows you to view and manage Knative Eventing resources.

The Backstage plugin talks to a special backend that runs in the Kubernetes cluster and communicates with the Kubernetes
API server.

A demo setup for this plugin is available at https://github.com/aliok/knative-backstage-demo.

## Installation

Install the backend and the relevant configuration in the Kubernetes cluster

```bash
kubectl apply -f https://github.com/knative-extensions/backstage-plugins/releases/download/v0.1.0/eventmesh.yaml
```

In your Backstage directory, run the following command to install the plugin:

```bash
yarn workspace backend add @knative-extensions/plugin-knative-event-mesh-backend
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
        baseUrl: 'http://localhost:8080' # URL of the backend installed in the cluster
        schedule: # optional; same options as in TaskScheduleDefinition
          # supports cron, ISO duration, "human duration" as used in code
          frequency: { minutes: 1 }
          # supports ISO duration, "human duration" as used in code
          timeout: { minutes: 1 }
```

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
