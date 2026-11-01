# Knative Backstage Plugins and Templates

## End of life

This project is being retired and this repository will be archived. It will no longer receive maintenance or new releases. Existing source code and release assets will remain available.

See [knative/community#1802](https://github.com/knative/community/issues/1802) for details.

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
