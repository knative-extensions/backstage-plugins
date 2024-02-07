import {getVoidLogger} from '@backstage/backend-common';
import {CatalogClient} from "@backstage/catalog-client";
import {ApiEntity, Entity} from "@backstage/catalog-model";
import {CatalogProcessorRelationResult} from "@backstage/plugin-catalog-node";
import {KnativeEventMeshProcessor} from "./knativeEventMeshProcessor";

// there must be a better way to do this
const catalogApi = <any>{
    queryEntities: jest.fn(),
} as jest.Mocked<CatalogClient>;

beforeEach(() => {
    catalogApi.queryEntities.mockClear();
});

describe('KnativeEventMeshProcessor', () => {
    const logger = getVoidLogger();

    describe('preProcessEntity', () => {

        const processor = new KnativeEventMeshProcessor(catalogApi, logger);

        type TestCase = {
            name:string;
            entity:ApiEntity;
            query?:{
                queryEntitiesRequest:{
                    filter:{
                        kind:'component',
                        'metadata.namespace':string,
                        'metadata.annotations.backstage.io/kubernetes-id':string,
                    },
                },
                queryEntitiesResult:Entity[];
            };
            expectedRelations?:CatalogProcessorRelationResult[];
        };

        const testCases:TestCase[] = [
            {
                name: 'should emit relations if consumer is defined and found',
                entity: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'API',
                    metadata: {
                        namespace: 'default',
                        name: 'et-1',
                        consumedBy: ['consumer-1'],
                    },
                    spec: {
                        owner: 'owner',
                        system: 'system',
                        lifecycle: 'lifecycle',
                        definition: 'definition',
                        type: 'eventType',
                    },
                },
                query: {
                    queryEntitiesRequest: {
                        filter: {
                            kind: 'component',
                            'metadata.namespace': 'default',
                            'metadata.annotations.backstage.io/kubernetes-id': 'consumer-1',
                        },
                    },
                    queryEntitiesResult: [{
                        apiVersion: 'backstage.io/v1alpha1',
                        kind: 'component',
                        metadata: {
                            namespace: 'default',
                            name: 'consumer-1',
                        },
                    }],
                },
                expectedRelations: [
                    {
                        type: 'relation',
                        relation: {
                            type: 'apiConsumedBy',
                            source: {
                                kind: 'API',
                                namespace: 'default',
                                name: 'et-1',
                            },
                            target: {
                                kind: 'Component',
                                namespace: 'default',
                                name: 'consumer-1',
                            },
                        },
                    },
                    {
                        type: 'relation',
                        relation: {
                            type: 'consumesApi',
                            source: {
                                kind: 'Component',
                                namespace: 'default',
                                name: 'consumer-1',
                            },
                            target: {
                                kind: 'API',
                                namespace: 'default',
                                name: 'et-1',
                            },
                        },
                    },
                ],
            },
            {
                "name": "should not emit relations if entity is not Knative Event Type",
                entity: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'API',
                    metadata: {
                        namespace: 'default',
                        name: 'et-1',
                    },
                    spec: {
                        owner: 'owner',
                        system: 'system',
                        lifecycle: 'lifecycle',
                        definition: 'definition',
                        type: 'RANDOM',
                    },
                },
            },
            {
                "name": "should not emit relations if there's no consumer defined",
                entity: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'API',
                    metadata: {
                        namespace: 'default',
                        name: 'et-1',
                        consumedBy: [],
                    },
                    spec: {
                        owner: 'owner',
                        system: 'system',
                        lifecycle: 'lifecycle',
                        definition: 'definition',
                        type: 'eventType',
                    },
                },
            },
            {
                name: 'should not emit relations if consumer is defined but cannot be found',
                entity: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'API',
                    metadata: {
                        namespace: 'default',
                        name: 'et-1',
                        consumedBy: ['consumer-1'],
                    },
                    spec: {
                        owner: 'owner',
                        system: 'system',
                        lifecycle: 'lifecycle',
                        definition: 'definition',
                        type: 'eventType',
                    },
                },
                query: {
                    queryEntitiesRequest: {
                        filter: {
                            kind: 'component',
                            'metadata.namespace': 'default',
                            'metadata.annotations.backstage.io/kubernetes-id': 'consumer-1',
                        },
                    },
                    queryEntitiesResult: [],
                },
                expectedRelations: [],
            },
        ];

        for (const testCase of testCases) {
            it(testCase.name, async () => {
                if (testCase.query) {
                    let entityQueryResult = {
                        items: testCase.query.queryEntitiesResult,
                        totalItems: testCase.query.queryEntitiesResult.length,
                        pageInfo: {}
                    };

                    catalogApi.queryEntities.mockReturnValue(Promise.resolve(entityQueryResult));
                }

                const emitFn = jest.fn();

                let output = await processor.preProcessEntity(testCase.entity, <any>{}, emitFn, <any>{}, <any>{});

                if (!testCase.expectedRelations) {
                    expect(emitFn).not.toHaveBeenCalled();
                } else {
                    expect(emitFn).toHaveBeenCalledTimes(testCase.expectedRelations.length);
                    for (let i = 0; i < testCase.expectedRelations.length; i++) {
                        const relation = testCase.expectedRelations[i];
                        expect(emitFn).toHaveBeenNthCalledWith(i + 1, relation);
                    }
                }

                expect(output).toEqual(testCase.entity);

                if (testCase.query) {
                    expect(catalogApi.queryEntities).toHaveBeenCalledTimes(1);
                    expect(catalogApi.queryEntities).toHaveBeenCalledWith(testCase.query.queryEntitiesRequest);
                } else {
                    expect(catalogApi.queryEntities).not.toHaveBeenCalled();
                }
            });
        }
    });
});
