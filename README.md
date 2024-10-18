# Knative Backstage Plugins and Templates

This repository contains a set of Backstage plugins for Knative, their respective backends and Knative Functions templates for Backstage.

## Installation and usage

### Event Mesh plugin

See [Event Mesh plugin README file](./backstage/plugins/knative-event-mesh-backend/README.md) for more information.

### Knative Function templates

See [templates README file](./backstage/templates/README.md) for more information.

## Development

### Event Mesh plugin

See [Event Mesh plugin DEVELOPMENT file](./backstage/plugins/knative-event-mesh-backend/DEVELOPMENT.md) for more information.

### Knative Function templates

See [templates README file](./backstage/templates/README.md) for more information.

### Testing GitHub Actions

You need `act` installed: https://github.com/nektos/act

```bash

# Specify the job to run
act -j '<job name>'
# ex:
# act -j 'publish-release-snapshot-on-npm'
# if having issues on Apple Silicon, use:
# act --rm --container-architecture linux/amd64 -j 'publish-release-snapshot-on-npm'
```
