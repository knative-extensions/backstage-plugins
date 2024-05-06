import {getVoidLogger} from '@backstage/backend-common';
import {ApiEntity, ComponentEntity} from "@backstage/catalog-model";

import {Broker, EventType, KnativeEventMeshProvider} from "./knativeEventMeshProvider";
import {KnativeEventMeshProviderConfig} from "./types";

describe('KnativeEventMeshProvider', () => {
    const logger = getVoidLogger();

    describe('buildEventTypeEntity', () => {
        const providerConfig:KnativeEventMeshProviderConfig = {
            id: 'test',
            baseUrl: 'http://example.com',
            schedule: undefined,
        };

        const provider = new KnativeEventMeshProvider(providerConfig, logger, (null as any));

        type TestCase = {
            name:string;
            input:EventType;
            expected:ApiEntity;
        };

        const testCases:TestCase[] = [
            {
                name: 'minimal information',
                input: {
                    name: 'test-name',
                    namespace: 'test-ns',
                    type: 'test-type',
                    uid: 'test-uid',
                },
                expected: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'API',
                    metadata: {
                        name: 'test-name',
                        namespace: 'test-ns',
                        title: 'test-type - (test-ns/test-name)',
                        description: undefined,
                        annotations: {
                            "backstage.io/managed-by-location": "url:http://example.com",
                            "backstage.io/managed-by-origin-location": "url:http://example.com",
                        },
                        labels: {},
                        tags: [],
                        links: [],
                        uid: undefined,
                        etag: undefined,
                        consumedBy: [],
                    },
                    spec: {
                        type: 'eventType',
                        owner: 'knative',
                        system: 'knative-event-mesh',
                        lifecycle: 'test',
                        definition: '{}',
                    },
                },
            },
            {
                name: 'all information',
                input: {
                    name: 'test-name',
                    namespace: 'test-ns',
                    type: 'test-type',
                    uid: 'test-uid',
                    description: 'test-description',
                    labels: {
                        "test-label": "test-label-value",
                    },
                    annotations: {
                        "test-annotation": "test-annotation-value",
                    },
                    schemaData: 'test-schema-data',
                    schemaURL: 'http://test-schema-url',
                    consumedBy: ["test-consumer1", "test-consumer2"],
                },
                expected: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'API',
                    metadata: {
                        name: 'test-name',
                        namespace: 'test-ns',
                        title: 'test-type - (test-ns/test-name)',
                        description: 'test-description',
                        annotations: {
                            "test-annotation": "test-annotation-value",
                            "backstage.io/managed-by-location": "url:http://example.com",
                            "backstage.io/managed-by-origin-location": "url:http://example.com",
                        },
                        labels: {
                            "test-label": "test-label-value",
                        },
                        tags: [],
                        links: [
                            {
                                title: "View external schema",
                                icon: "scaffolder",
                                url: 'http://test-schema-url',
                            },
                        ],
                        uid: undefined,
                        etag: undefined,
                        consumedBy: ["test-consumer1", "test-consumer2"],
                    },
                    spec: {
                        type: 'eventType',
                        owner: 'knative',
                        system: 'knative-event-mesh',
                        lifecycle: 'test',
                        definition: 'test-schema-data',
                    },
                },
            }
        ];

        for (const testCase of testCases) {
            test(`Name: ${testCase.name}`, async () => {
                const result = provider.buildEventTypeEntity(testCase.input);
                expect(result).toEqual(testCase.expected);
            });
        }
    });

    describe('buildBrokerEntity', () => {
        const providerConfig:KnativeEventMeshProviderConfig = {
            id: 'test',
            baseUrl: 'http://example.com',
            schedule: undefined,
        };

        const provider = new KnativeEventMeshProvider(providerConfig, logger, (null as any));

        type TestCase = {
            name:string;
            input:Broker;
            expected:ComponentEntity;
        };

        const testCases:TestCase[] = [
            {
                name: 'minimal information',
                input: {
                    name: 'test',
                    namespace: 'test',
                    uid: 'test',
                },
                expected: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'Component',
                    metadata: {
                        name: 'test',
                        namespace: 'test',
                        annotations: {
                            "backstage.io/managed-by-location": "url:http://example.com",
                            "backstage.io/managed-by-origin-location": "url:http://example.com",
                        },
                        labels: {},
                        tags: [],
                    },
                    spec: {
                        type: 'broker',
                        owner: 'knative',
                        system: 'knative-event-mesh',
                        lifecycle: 'test',
                        providesApis: [],
                    },
                },
            },
            {
                name: 'all information',
                input: {
                    name: 'test-broker',
                    namespace: 'test-ns',
                    uid: 'test-uid',
                    labels: {
                        "test-label": "test-label-value",
                    },
                    annotations: {
                        "test-annotation": "test-annotation-value",
                    },
                    providedEventTypes: [
                        "test-ns/test-type",
                    ],
                },
                expected: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'Component',
                    metadata: {
                        name: 'test-broker',
                        namespace: 'test-ns',
                        annotations: {
                            "test-annotation": "test-annotation-value",
                            "backstage.io/managed-by-location": "url:http://example.com",
                            "backstage.io/managed-by-origin-location": "url:http://example.com",
                        },
                        labels: {
                            "test-label": "test-label-value",
                        },
                        tags: [],
                    },
                    spec: {
                        type: 'broker',
                        owner: 'knative',
                        system: 'knative-event-mesh',
                        lifecycle: 'test',
                        providesApis: [
                            "api:test-ns/test-type"
                        ],
                    },
                },
            }
        ];

        for (const testCase of testCases) {
            test(`Name: ${testCase.name}`, async () => {
                const result = provider.buildBrokerEntity(testCase.input);
                expect(result).toEqual(testCase.expected);
            });

        }
    });
});
