import {BackendDynamicPluginInstaller} from '@backstage/backend-dynamic-feature-service';
import {CatalogClient} from "@backstage/catalog-client";

import {KnativeEventMeshProcessor, KnativeEventMeshProvider} from '../providers';

// This is mainly to provide dynamic plugin support for backstage's legacy backend. During the plugin loading phase, it'll use this dynamicPluginInstaller.
// https://github.com/backstage/backstage/blob/master/packages/backend-dynamic-feature-service/src/manager/types.ts
// It's deprecated since the legacy backend is now deprecated, but it's probably still nice to have in case someone wants to install on the legacy backend for some reason.
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
