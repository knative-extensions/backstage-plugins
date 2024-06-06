import {loggerToWinstonLogger} from '@backstage/backend-common';
import {
    coreServices,
    createBackendModule,
} from '@backstage/backend-plugin-api';
import { 
    catalogServiceRef,
    catalogProcessingExtensionPoint
} from '@backstage/plugin-catalog-node/alpha';

import { 
    KnativeEventMeshProcessor, 
    KnativeEventMeshProvider
} from './providers';

export const catalogModuleKnativeEventMesh = createBackendModule({
    moduleId: 'knative-event-mesh-module',
    pluginId: 'catalog',
    register(env) {
        env.registerInit({
            deps: {
                catalogApi: catalogServiceRef,
                catalog: catalogProcessingExtensionPoint,
                config: coreServices.rootConfig,
                logger: coreServices.logger,
                scheduler: coreServices.scheduler,
            },
            async init({ catalogApi, catalog, config, logger, scheduler }) {
                const knativeEventMeshProviders = KnativeEventMeshProvider.fromConfig(config, {
                    logger: loggerToWinstonLogger(logger),
                    scheduler: scheduler,
                });
                catalog.addEntityProvider(knativeEventMeshProviders);

                const knativeEventMeshProcessor = new KnativeEventMeshProcessor(catalogApi, loggerToWinstonLogger(logger));
                catalog.addProcessor(knativeEventMeshProcessor);
            },
        });
    },
});
