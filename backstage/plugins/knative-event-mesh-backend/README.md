# Knative Event Mesh plugin

The Event Mesh plugin is a Backstage plugin that allows you to view and manage Knative Eventing resources.

A demo setup for this plugin is available at https://github.com/aliok/knative-backstage-demo.

## Dynamic vs static plugin

This plugin has 2 distributions: static and dynamic.

The static distribution is a regular Backstage plugin that requires the source code of Backstage to be changed.

The dynamic distribution is a plugin that can be installed without changing the source code of Backstage.

If you would like to use the dynamic plugin, please see the instructions in the
[Dynamic Plugin README file](https://github.com/knative-extensions/backstage-plugins/blob/main/backstage/plugins/knative-event-mesh-backend/README-dynamic.md).

If you would like to use the static distribution, please see the documentation on Knative website for
[installing](https://knative.dev/docs/install/installing-backstage-plugins/)
and [using](https://knative.dev/docs/eventing/event-registry/eventmesh-backstage-plugin/) the plugin.
