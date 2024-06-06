import {CatalogApi} from '@backstage/catalog-client';
import {ComponentEntity, Entity} from '@backstage/catalog-model';
import {LocationSpec} from '@backstage/plugin-catalog-common';
import {
    CatalogProcessor,
    CatalogProcessorCache,
    CatalogProcessorEmit,
    CatalogProcessorRelationResult,
} from '@backstage/plugin-catalog-node';
import {Logger} from "winston";
import {TypeKnativeEvent} from "./types";


export class KnativeEventMeshProcessor implements CatalogProcessor {
    private readonly catalogApi: CatalogApi;
    private readonly logger:Logger;
    private readonly queryEntityPageLimit:number;

    constructor(catalogApi:CatalogApi, logger:Logger, queryEntityPageLimit?:number) {
        this.catalogApi = catalogApi;
        this.queryEntityPageLimit = queryEntityPageLimit ?? 10000;

        this.logger = logger.child({
            target: this.getProcessorName(),
        });
    }

    getProcessorName():string {
        return "knative-event-mesh-processor";
    }

    async preProcessEntity(entity:Entity, _location:LocationSpec, emit:CatalogProcessorEmit, _originLocation:LocationSpec, _cache:CatalogProcessorCache):Promise<Entity> {
        if (entity.kind === 'API' && entity.spec?.type === TypeKnativeEvent) {
            this.logger.debug(`Processing KnativeEventType entity ${entity.metadata.namespace}/${entity.metadata.name}`);

            // if there's no relation to build, return entity as is
            if (!entity.metadata.consumedBy) {
                this.logger.debug(`No consumers defined for KnativeEventType entity ${entity.metadata.namespace}/${entity.metadata.name}`);
                return entity;
            }

            const consumers = entity.metadata.consumedBy as string[];
            this.logger.debug(`Consumers defined for KnativeEventType entity ${entity.metadata.namespace}/${entity.metadata.name}: ${consumers.join(', ')}`);

            // build relations
            for (const consumedBy of consumers) {
                this.logger.debug(`Building relations for KnativeEventType entity ${entity.metadata.namespace}/${entity.metadata.name} to consumer ${consumedBy}`);

                // query the catalog for the component with the id
                const consumerComponents = await this.findComponentsByBackstageId(entity.metadata.namespace as string, consumedBy);
                this.logger.debug(`Found ${consumerComponents.length} components for KnativeEventType entity ${entity.metadata.namespace}/${entity.metadata.name} to consumer ${consumedBy}`);

                for (const component of consumerComponents) {
                    this.logger.debug(`Emitting relations for KnativeEventType entity ${entity.metadata.namespace}/${entity.metadata.name} for consumer ${consumedBy} via component ${component.metadata.namespace}/${component.metadata.name}`);

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
        // fetch the component by the id
        // example: http://localhost:7007/api/catalog/entities/by-query
        // ?filter=kind=component,metadata.namespace=default,metadata.annotations.backstage.io/kubernetes-id=fraud-detector
        let catalogApiCursor: string | undefined;
        let entities: Entity[] = [];

        try {
            do {
                const response = await this.catalogApi.queryEntities({
                    filter: {
                        kind: 'component',
                        'metadata.namespace': namespace,
                        'metadata.annotations.backstage.io/kubernetes-id': componentId,
                    },
                    cursor: catalogApiCursor,
                    limit: this.queryEntityPageLimit
                });
                catalogApiCursor = response.pageInfo.nextCursor;
                entities = entities.concat(response.items);
            } while (catalogApiCursor)

            return entities;
        } catch (e) {
            this.logger.error(`Failed to find components by backstage id ${namespace}/${componentId}: ${e}`);
            return [] as ComponentEntity[];
        }
    }
}
