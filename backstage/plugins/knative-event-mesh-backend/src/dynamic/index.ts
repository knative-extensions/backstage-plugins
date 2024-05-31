import {BackendDynamicPluginInstaller} from '@backstage/backend-dynamic-feature-service';
import {CatalogClient} from "@backstage/catalog-client";

import {KnativeEventMeshProcessor, KnativeEventMeshProvider} from '../providers';

export const dynamicPluginInstaller:BackendDynamicPluginInstaller = {
    kind: 'legacy',
    async catalog(builder:any, env:any) {
        const knativeEventMeshProviders = KnativeEventMeshProvider.fromConfig(env.config, {
            logger: env.logger,
            scheduler: env.scheduler,
        });
        builder.addEntityProvider(knativeEventMeshProviders);

        const catalogApi = new CatalogClient({
            discoveryApi: env.discovery,
        });

        const knativeEventMeshProcessor = new KnativeEventMeshProcessor(catalogApi, env.logger);
        builder.addProcessor(knativeEventMeshProcessor);
    },
};
