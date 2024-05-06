import {getVoidLogger} from '@backstage/backend-common';
import {CatalogClient} from '@backstage/catalog-client';
import {ApiEntity, Entity} from '@backstage/catalog-model';
import {CatalogProcessorRelationResult} from '@backstage/plugin-catalog-node';
import {KnativeEventMeshProcessor} from "./knativeEventMeshProcessor";

// there must be a better way to do this
const catalogApi = ({
    queryEntities: jest.fn(),
} as any) as jest.Mocked<CatalogClient>;

beforeEach(() => {
    catalogApi.queryEntities.mockClear();
});

describe('KnativeEventMeshProcessor', () => {
    const logger = getVoidLogger();

    describe('preProcessEntity', () => {

        const processor = new KnativeEventMeshProcessor(catalogApi, logger, 5);

        type Query = {
            queryEntitiesRequest:{
                filter:{
                    kind:'component',
                    'metadata.namespace':string,
                    'metadata.annotations.backstage.io/kubernetes-id':string,
                },
                cursor: string|undefined;
                limit: number
            },
            queryEntitiesResult: {
               items: Entity[],
                pageInfo: {
                    nextCursor?: string;
                    prevCursor?: string;
                };
            };
        }

        interface TestCase  {
            name:string;
            entity:ApiEntity;
            queries?:Query[];
            expectedRelations?:CatalogProcessorRelationResult[];
        }

        const testCases:TestCase[] = [
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
                queries: [{
                    queryEntitiesRequest: {
                        filter: {
                            kind: 'component',
                            'metadata.namespace': 'default',
                            'metadata.annotations.backstage.io/kubernetes-id': 'consumer-1',
                        },
                        cursor:undefined,
                        limit: 5
                    },
                    queryEntitiesResult: {
                        items: [],
                        pageInfo: {

                        },
                    }
                }],
                expectedRelations: [
                ],
            },
            {
                name: 'should make a 2nd call with cursor if there are more than 5 items returned and limit is 5',
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
                queries: [
                    {
                    queryEntitiesRequest: {
                        filter: {
                            kind: 'component',
                            'metadata.namespace': 'default',
                            'metadata.annotations.backstage.io/kubernetes-id': 'consumer-1',
                        },
                        cursor: undefined,
                        limit: 5,
                    },
                    queryEntitiesResult:
                        {
                          items: [
                                {
                                    apiVersion: 'backstage.io/v1alpha1',
                                    kind: 'component',
                                    metadata: {
                                        namespace: 'default',
                                        name: 'consumer-1',
                                    },
                                },
                                {
                                    apiVersion: 'backstage.io/v1alpha',
                                    kind: 'component',
                                    metadata: {
                                        namespace: 'default',
                                        name: 'consumer-1',
                                    },
                                },
                                {
                                    apiVersion: 'backstage.io/v1alpha1',
                                    kind: 'component',
                                    metadata: {
                                        namespace: 'default',
                                        name: 'consumer-1',
                                    },
                                },
                                {
                                    apiVersion: 'backstage.io/v1alpha1',
                                    kind: 'component',
                                    metadata: {
                                        namespace: 'default',
                                        name: 'consumer-1',
                                    },
                                },
                                {
                                    apiVersion: 'backstage.io/v1alpha1',
                                    kind: 'component',
                                    metadata: {
                                        namespace: 'default',
                                        name: 'consumer-1',
                                    },
                                },
                            ],
                          pageInfo: {
                                nextCursor: "2",
                                prevCursor: "1"
                           }
                        }
                    },
                    {
                        queryEntitiesRequest: {
                            filter: {
                                kind: 'component',
                                'metadata.namespace': 'default',
                                'metadata.annotations.backstage.io/kubernetes-id': 'consumer-1',
                            },
                            cursor: "2",
                            limit: 5,
                        },
                        queryEntitiesResult:
                            {
                              items: [
                                    {
                                        apiVersion: 'backstage.io/v1alpha1',
                                        kind: 'component',
                                        metadata: {
                                            namespace: 'default',
                                            name: 'consumer-1',
                                        },
                                    },
                                    {
                                        apiVersion: 'backstage.io/v1alpha2',
                                        kind: 'component',
                                        metadata: {
                                            namespace: 'default',
                                            name: 'consumer-1',
                                        },
                                    },
                                ],
                              pageInfo: {
                                    nextCursor: undefined,
                                    prevCursor: "1"
                               }
                            }
                        },
                ],
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
                    }
                ],
            }
        ];

        for (const testCase of testCases) {
            test(`Name: ${testCase.name}`, async () => {
                if (testCase.queries) {
                    for (const query of testCase.queries) {
                        const entityQueryResult = {
                            items: query.queryEntitiesResult.items,
                            totalItems: query.queryEntitiesResult.items.length,
                            pageInfo: query.queryEntitiesResult.pageInfo
                        };
                        catalogApi.queryEntities.mockReturnValueOnce(Promise.resolve(entityQueryResult));
                    }
                }

                const emitFn = jest.fn();

                const output = await processor.preProcessEntity(testCase.entity, ({} as any), emitFn, ({} as any), ({} as any));

                expect(emitFn).toHaveBeenCalledTimes(testCase.expectedRelations?.length || 0);

                testCase.expectedRelations?.forEach((relation, index) => {
                    expect(emitFn).toHaveBeenNthCalledWith(index + 1, relation);
                });

                expect(output).toEqual(testCase.entity);

                expect(catalogApi.queryEntities).toHaveBeenCalledTimes(testCase.queries?.length || 0);

                testCase.queries?.forEach(query => {
                    expect(catalogApi.queryEntities).toHaveBeenCalledWith(query.queryEntitiesRequest);
                });
            });
        }

    });
});
