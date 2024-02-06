import {Entity} from '@backstage/catalog-model';
import {LocationSpec} from '@backstage/plugin-catalog-common';
import {
    CatalogProcessor,
    CatalogProcessorCache,
    CatalogProcessorEmit,
    CatalogProcessorRelationResult,
} from '@backstage/plugin-catalog-node';


export class KnativeEventMeshProcessor implements CatalogProcessor {
    getProcessorName():string {
        // TODO: append env?
        return `knative-event-mesh-processor`;
    }

    async preProcessEntity(entity:Entity, _location:LocationSpec, emit:CatalogProcessorEmit, _originLocation:LocationSpec, _cache:CatalogProcessorCache):Promise<Entity> {
        // TODO: remove hardcoded strings
        if (entity.kind === 'API' && entity.spec?.type === 'eventType') {
            // if there's no relation to build, return entity as is
            if (!entity.metadata.consumedBy) {
                return entity;
            }

            const consumers = entity.metadata.consumedBy as string[];

            // build relations
            for (const consumedBy of consumers) {
                // TODO: query the catalog for the component with the id

                // emit a relation from the API to the component
                const apiToComponentRelation:CatalogProcessorRelationResult = {
                    type: 'relation',
                    relation: {
                        type: 'apiConsumedBy',
                        source: {
                            kind: 'API',
                            namespace: entity.metadata.namespace as string,
                            name: entity.metadata.name,
                        },
                        target: {
                            kind: 'Component',
                            namespace: entity.metadata.namespace as string,
                            name: consumedBy,
                        },
                    },
                };
                emit(apiToComponentRelation);

                // emit a relation from the component to the API
                const componentToApiRelation:CatalogProcessorRelationResult = {
                    type: 'relation',
                    relation: {
                        type: 'consumesApi',
                        source: {
                            kind: 'Component',
                            namespace: entity.metadata.namespace as string,
                            name: consumedBy,
                        },
                        target: {
                            kind: 'API',
                            namespace: entity.metadata.namespace as string,
                            name: entity.metadata.name,
                        },
                    },
                };
                emit(componentToApiRelation);
            }
        }
        return entity;
    }
}
