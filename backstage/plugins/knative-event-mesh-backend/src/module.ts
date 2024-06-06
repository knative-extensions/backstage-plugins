import {loggerToWinstonLogger} from '@backstage/backend-common';
import {
    coreServices,
    createBackendModule,
} from '@backstage/backend-plugin-api';
import {CatalogClient} from "@backstage/catalog-client";
import {catalogProcessingExtensionPoint} from '@backstage/plugin-catalog-node/alpha';

import {KnativeEventMeshProcessor, KnativeEventMeshProvider} from './providers';

export const catalogModuleKnativeEventMesh = createBackendModule({
    moduleId: 'knative-event-mesh-module',
    pluginId: 'catalog',
    register(env) {
        env.registerInit({
            deps: {
                catalog: catalogProcessingExtensionPoint,
                config: coreServices.rootConfig,
                logger: coreServices.logger,
                scheduler: coreServices.scheduler,
                discovery: coreServices.discovery,
            },
            async init({catalog, config, logger, scheduler, discovery}) {
                const knativeEventMeshProviders = KnativeEventMeshProvider.fromConfig(config, {
                    logger: loggerToWinstonLogger(logger),
                    scheduler: scheduler,
                });
                catalog.addEntityProvider(knativeEventMeshProviders);

                const catalogApi = new CatalogClient({
                    discoveryApi: discovery,
                });

                const knativeEventMeshProcessor = new KnativeEventMeshProcessor(catalogApi, loggerToWinstonLogger(logger));
                catalog.addProcessor(knativeEventMeshProcessor);
            },
        });
    },
});
