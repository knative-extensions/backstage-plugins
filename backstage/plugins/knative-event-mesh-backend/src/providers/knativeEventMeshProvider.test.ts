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

        const provider = new KnativeEventMeshProvider(providerConfig, logger, <any>null);

        type TestCase = {
            name:string;
            input:EventType;
            expected:ApiEntity;
        };

        const testCases:TestCase[] = [
            {
                name: 'minimal information',
                input: {
                    name: 'test',
                    namespace: 'test',
                    type: 'test',
                    uid: 'test',
                },
                expected: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'API',
                    metadata: {
                        name: 'test',
                        namespace: 'test',
                        description: undefined,
                        annotations: {
                            "backstage.io/managed-by-location": "url:http://example.com/eventtype/test/test",
                            "backstage.io/managed-by-origin-location": "url:http://example.com/eventtype/test/test",
                        },
                        labels: {},
                        tags: [],
                        links: [],
                        uid: undefined,
                        title: undefined,
                        etag: undefined,
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
                    name: 'test',
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
                },
                expected: {
                    apiVersion: 'backstage.io/v1alpha1',
                    kind: 'API',
                    metadata: {
                        name: 'test-type',
                        namespace: 'test-ns',
                        description: 'test-description',
                        annotations: {
                            "test-annotation": "test-annotation-value",
                            "backstage.io/managed-by-location": "url:http://example.com/eventtype/test-ns/test",
                            "backstage.io/managed-by-origin-location": "url:http://example.com/eventtype/test-ns/test",
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
                        title: undefined,
                        etag: undefined,
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
            it(testCase.name, async () => {
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

        const provider = new KnativeEventMeshProvider(providerConfig, logger, <any>null);

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
                            "backstage.io/managed-by-location": "url:http://example.com/broker/test/test",
                            "backstage.io/managed-by-origin-location": "url:http://example.com/broker/test/test",
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
                            "backstage.io/managed-by-location": "url:http://example.com/broker/test-ns/test-broker",
                            "backstage.io/managed-by-origin-location": "url:http://example.com/broker/test-ns/test-broker",
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
            it(testCase.name, async () => {
                const result = provider.buildBrokerEntity(testCase.input);
                expect(result).toEqual(testCase.expected);
            });
        }
    });
});
