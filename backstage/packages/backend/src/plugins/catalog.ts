import {CatalogClient} from "@backstage/catalog-client";
import { CatalogBuilder } from '@backstage/plugin-catalog-backend';
import { ScaffolderEntitiesProcessor } from '@backstage/plugin-catalog-backend-module-scaffolder-entity-model';
import { Router } from 'express';
import { PluginEnvironment } from '../types';
import {
  KnativeEventMeshProcessor,
  KnativeEventMeshProvider
} from '@knative-extensions/plugin-knative-event-mesh-backend';

export default async function createPlugin(
  env: PluginEnvironment,
): Promise<Router> {
  const builder = CatalogBuilder.create(env);
  builder.addProcessor(new ScaffolderEntitiesProcessor());

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

  const { processingEngine, router } = await builder.build();
  await processingEngine.start();

  return router;
}
