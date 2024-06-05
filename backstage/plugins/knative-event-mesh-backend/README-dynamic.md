# Knative Event Mesh plugin (dynamic)

The Event Mesh plugin is a Backstage plugin that allows you to view and manage Knative Eventing resources.

The Backstage plugin talks to a special backend that runs in the Kubernetes cluster and communicates with the Kubernetes
API server.

A demo setup for this plugin is available at https://github.com/aliok/knative-backstage-demo.

For more information about the plugin, please see
the [GitHub repository](https://github.com/knative-extensions/backstage-plugins/blob/main/backstage/plugins/knative-event-mesh-backend/README.md).

This distribution of the plugin is a dynamic plugin that can be installed in a Backstage instance that supports dynamic
plugins.

Backstage has a [WIP proposal](https://github.com/backstage/backstage/tree/master/beps/0002-dynamic-frontend-plugins) to
allow plugins to be loaded dynamically. While this is not done in upstream Backstage
yet, [Janus IDP](https://janus-idp.io/) has implemented this feature.
The benefit of the dynamic plugin is it can be used without changing the source code of Backstage.

## Installation

### Prerequisites

- A Kubernetes cluster with Knative Eventing installed
- Knative Event Mesh plugin backend installed
- A Backstage instance with dynamic plugin support (e.g. [Janus IDP](https://janus-idp.io/))
- A service account for the Backstage backend to access the Kubernetes API

Install Knative Eventing by following the [official documentation](https://knative.dev/docs/install/).

Install the backend and the relevant configuration in the Kubernetes cluster

```bash
kubectl apply -f https://github.com/knative-extensions/backstage-plugins/releases/download/v0.1.0/eventmesh.yaml
```

## Janus Configuration

You need to follow the Janus IDP dynamic plugin installation instructions
here: https://github.com/janus-idp/backstage-showcase/blob/main/showcase-docs/dynamic-plugins.md#installing-a-dynamic-plugin-package-in-the-showcase

For a quick test, download the plugin package and extract it to the `dynamic-plugins-root` directory in Janus IDP:

```bash
cd <path-to-Janus-IDP>/dynamic-plugins-root
pkg=@knative-extensions/plugin-knative-event-mesh-backend-dynamic
pkg=@knative-extensions/plugin-knative-event-mesh-backend-dynamic
pkg=@aliok/plugin-knative-event-mesh-backend-dynamic
archive=$(npm pack $pkg)
tar -xzf "$archive" && rm "$archive"
mv package $(echo $archive | sed -e 's:\.tgz$::')
```

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
  ...
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

Start your Janus IDP instance!

> **NOTE**: If you have made any changes to the schedule in the `app-config.yaml` file, then restart to apply the
> changes.

## Troubleshooting

When you start your Backstage application, you can see some log lines as follows:

```text
[1] 2024-01-04T09:38:08.707Z knative-event-mesh-backend info Found 1 knative event mesh provider configs with ids: dev type=plugin
```
