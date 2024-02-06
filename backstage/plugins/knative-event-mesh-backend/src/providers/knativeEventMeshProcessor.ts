import {CatalogClient} from '@backstage/catalog-client';
import {ComponentEntity, Entity} from '@backstage/catalog-model';
import {LocationSpec} from '@backstage/plugin-catalog-common';
import {
    CatalogProcessor,
    CatalogProcessorCache,
    CatalogProcessorEmit,
    CatalogProcessorRelationResult,
} from '@backstage/plugin-catalog-node';


export class KnativeEventMeshProcessor implements CatalogProcessor {
    private catalogApi:CatalogClient;

    constructor(catalogApi:CatalogClient) {
        this.catalogApi = catalogApi;
    }

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
                // query the catalog for the component with the id
                const consumerComponents = await this.findComponentsByBackstageId(entity.metadata.namespace as string, consumedBy);

                for (const component of consumerComponents) {
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
                                namespace: component.metadata.namespace as string,
                                name: component.metadata.name,
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
                                namespace: component.metadata.namespace as string,
                                name: component.metadata.name,
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
        }
        return entity;
    }

    private async findComponentsByBackstageId(namespace:string, componentId:string) {
        // TODO: make use of `backstage.io/kubernetes-namespace` annotation?
        // TODO: make use of `backstage.io/kubernetes-label-selector` annotation?

        // fetch the component by the id
        // example: http://localhost:7007/api/catalog/entities/by-query?filter=kind=component,metadata.namespace=default,metadata.annotations.backstage.io/kubernetes-id=fraud-detector

        try {
            let response = await this.catalogApi.queryEntities({
                filter: {
                    kind: 'component',
                    'metadata.namespace': namespace,
                    // TODO: hardcoded annotation
                    'metadata.annotations.backstage.io/kubernetes-id': componentId,
                },
            });

            return response.items as ComponentEntity[];
        } catch (e) {
            // TODO: log
            return [] as ComponentEntity[];
        }
    }
}
