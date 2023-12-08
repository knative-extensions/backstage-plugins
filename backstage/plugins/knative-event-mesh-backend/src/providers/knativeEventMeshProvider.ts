import {PluginTaskScheduler, TaskRunner} from '@backstage/backend-tasks';
import {
    ApiEntity,
    Entity,
    ANNOTATION_LOCATION,
    ANNOTATION_ORIGIN_LOCATION,
    EntityLink, ComponentEntity,
} from '@backstage/catalog-model';

import {Config} from '@backstage/config';

import {EntityProvider, EntityProviderConnection,} from '@backstage/plugin-catalog-node';

import {Logger} from 'winston';
import {readKnativeEventMeshProviderConfigs} from "./config";
import {KnativeEventMeshProviderConfig} from "./types";

type EventType = {
    name:string;
    namespace:string;
    type:string;
    uid:string;
    description?:string;
    schemaData?:string;
    schemaURL?:string;
    labels?:Record<string, string>;
    annotations?:Record<string, string>;
};

type Broker = {
    name:string;
    namespace:string;
    uid:string;
    labels?:Record<string, string>;
    annotations?:Record<string, string>;
    providedEventTypes?:string[];
};

type EventMesh = {
    eventTypes:EventType[];
    brokers:Broker[];
};

export async function getEventMesh(baseUrl:string):Promise<EventMesh> {
    const response = await fetch(`${baseUrl}`);
    if (!response.ok) {
        throw new Error(response.statusText);
    }
    return await response.json() as Promise<EventMesh>;
}

export class KnativeEventMeshProvider implements EntityProvider {
    private readonly env:string;
    private readonly baseUrl:string;
    private readonly logger:Logger;
    private readonly scheduleFn:() => Promise<void>;
    private connection?:EntityProviderConnection;

    static fromConfig(
        configRoot:Config,
        options:{
            logger:Logger;
            schedule?:TaskRunner;
            scheduler?:PluginTaskScheduler;
        },
    ):KnativeEventMeshProvider[] {
        const providerConfigs = readKnativeEventMeshProviderConfigs(configRoot);

        if (!options.schedule && !options.scheduler) {
            throw new Error('Either schedule or scheduler must be provided.');
        }

        const logger = options.logger.child({plugin: 'knative-event-mesh-backend'});
        logger.info(`Found ${providerConfigs.length} knative event mesh provider configs with ids: ${providerConfigs.map(providerConfig => providerConfig.id).join(', ')}`);

        return providerConfigs.map(providerConfig => {
            if (!options.schedule && !providerConfig.schedule) {
                throw new Error(`No schedule provided neither via code nor config for KnativeEventMesh entity provider:${providerConfig.id}.`);
            }

            let taskRunner;

            if (options.scheduler && providerConfig.schedule) {
                // Create a scheduled task runner using the provided scheduler and schedule configuration
                taskRunner = options.scheduler.createScheduledTaskRunner(providerConfig.schedule);
            } else if (options.schedule) {
                // Use the provided schedule directly
                taskRunner = options.schedule;
            } else {
                // Handle the case where both options.schedule and options.scheduler are missing
                throw new Error('Neither schedule nor scheduler is provided.');
            }

            return new KnativeEventMeshProvider(
                providerConfig,
                options.logger,
                taskRunner,
            );
        });
    }

    constructor(config:KnativeEventMeshProviderConfig, logger:Logger, taskRunner:TaskRunner) {
        this.env = config.id;
        this.baseUrl = config.baseUrl;

        this.logger = logger.child({
            target: this.getProviderName(),
        });

        this.scheduleFn = this.createScheduleFn(taskRunner);
    }

    private createScheduleFn(taskRunner:TaskRunner):() => Promise<void> {
        return async () => {
            const taskId = `${this.getProviderName()}:run`;
            return taskRunner.run({
                id: taskId,
                fn: async () => {
                    try {
                        await this.run();
                    } catch (error:any) {
                        // Ensure that we don't log any sensitive internal data:
                        this.logger.error(
                            `Error while fetching Knative Event Mesh from ${this.baseUrl}`,
                            {
                                // Default Error properties:
                                name: error.name,
                                message: error.message,
                                stack: error.stack,
                                // Additional status code if available:
                                status: error.response?.status,
                            },
                        );
                    }
                },
            });
        };
    }

    getProviderName():string {
        return `knative-event-mesh-${this.env}`;
    }

    async connect(connection:EntityProviderConnection):Promise<void> {
        this.connection = connection;
        await this.scheduleFn();
    }

    async run():Promise<void> {
        if (!this.connection) {
            throw new Error('Not initialized');
        }

        const eventMesh = await getEventMesh(this.baseUrl);

        const entities:Entity[] = [];

        for (const eventType of eventMesh.eventTypes) {
            const entity = this.buildEventTypeEntity(eventType);
            entities.push(entity);
        }

        for (const broker of eventMesh.brokers) {
            const entity = this.buildBrokerEntity(broker);
            entities.push(entity);
        }

        await this.connection.applyMutation({
            type: 'full',
            entities: entities.map(entity => ({
                entity,
                locationKey: this.getProviderName(),
            })),
        });
    }

    private buildEventTypeEntity(eventType:EventType):ApiEntity {
        const annotations = eventType.annotations ?? {} as Record<string, string>;
        // TODO: no route exists yet
        annotations[ANNOTATION_ORIGIN_LOCATION] = annotations[ANNOTATION_LOCATION] = `url:${this.baseUrl}/eventtype/${eventType.namespace}/${eventType.name}`;

        const links:EntityLink[] = [];
        if (eventType.schemaURL) {
            links.push({
                title: "View external schema",
                icon: "scaffolder",
                url: eventType.schemaURL
            });
        }

        // TODO: remove?
        // let relations:EntityRelation[] = [];
        // if (eventType.provider) {
        //     relations = [...relations, {
        //         // type: RELATION_API_PROVIDED_BY,
        //         type: 'apiProvidedBy',
        //         // TODO: ref should point to the Backstage Broker provider
        //         // targetRef: `${this.getProviderName()}:${eventType.provider.kind}:${eventType.provider.namespace}/${eventType.provider.name}`,
        //         targetRef: `component:default/example-website`,
        //     }];
        //     console.log(relations);
        //
        //     // TODO:
        //     // partOf: https://backstage.io/docs/features/software-catalog/well-known-relations/#partof-and-haspart
        //     // - system?
        //
        //     // TODO:
        //     // apiConsumedBy: https://backstage.io/docs/features/software-catalog/well-known-relations/#consumesapi-and-apiconsumedby
        //     // - triggers?
        // }

        return {
            apiVersion: 'backstage.io/v1alpha1',
            kind: 'API',
            metadata: {
                name: eventType.type,
                namespace: eventType.namespace,
                description: eventType.description,
                // TODO: is there a value showing Kubernetes labels in Backstage?
                labels: eventType.labels || {} as Record<string, string>,
                // TODO: is there a value showing Kubernetes annotations in Backstage?
                annotations: annotations,
                // we don't use tags
                tags: [],
                links: links,
            },
            spec: {
                type: 'eventType',
                lifecycle: this.env,
                // TODO
                system: 'knative-event-mesh',
                // TODO
                owner: 'knative',
                definition: eventType.schemaData || "{}",
            },
            // TODO: remove?
            // Backstage doesn't like empty relations
            // relations: relations.length > 0 ? relations : undefined,
        };
    }

    private buildBrokerEntity(broker:Broker): ComponentEntity {
        const annotations = broker.annotations ?? {} as Record<string, string>;
        // TODO: no route exists yet
        annotations[ANNOTATION_ORIGIN_LOCATION] = annotations[ANNOTATION_LOCATION] = `url:${this.baseUrl}/broker/${broker.namespace}/${broker.name}`;

        return {
            apiVersion: 'backstage.io/v1alpha1',
            kind: 'Component',
            metadata: {
                // TODO: names are too generic: default/default
                name: broker.name,
                namespace: broker.namespace,
                // TODO: is there a value showing Kubernetes labels in Backstage?
                labels: broker.labels || {} as Record<string, string>,
                // TODO: is there a value showing Kubernetes annotations in Backstage?
                annotations: annotations,
                // we don't use tags
                tags: [],
                // TODO: any links?
                // links: links,
            },
            spec: {
                type: 'broker',
                lifecycle: this.env,
                // TODO
                system: 'knative-event-mesh',
                // TODO
                owner: 'knative',
                providesApis: !broker.providedEventTypes ? [] : broker.providedEventTypes.map((eventType:string) => `api:${eventType}`),
            }
        }
    }
}
