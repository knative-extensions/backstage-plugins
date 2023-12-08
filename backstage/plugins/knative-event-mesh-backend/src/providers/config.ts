import {readTaskScheduleDefinitionFromConfig} from '@backstage/backend-tasks';
import {Config} from '@backstage/config';

import {KnativeEventMeshProviderConfig} from './types';

export function readKnativeEventMeshProviderConfigs(config:Config):KnativeEventMeshProviderConfig[] {
    const providerConfigs = config.getOptionalConfig(
        'catalog.providers.knativeEventMesh',
    );
    if (!providerConfigs) {
        return [];
    }
    return providerConfigs
        .keys()
        .map(id =>
            readKnativeEventMeshProviderConfig(id, providerConfigs.getConfig(id)),
        );
}

function readKnativeEventMeshProviderConfig(id:string, config:Config):KnativeEventMeshProviderConfig {
    const baseUrl = config.getString('baseUrl');

    const schedule = config.has('schedule')
        ? readTaskScheduleDefinitionFromConfig(config.getConfig('schedule'))
        : undefined;

    return {
        id,
        baseUrl,
        schedule,
    };
}
