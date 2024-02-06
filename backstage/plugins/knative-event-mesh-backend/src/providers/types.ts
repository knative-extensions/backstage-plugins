import {TaskScheduleDefinition} from '@backstage/backend-tasks';

export type KnativeEventMeshProviderConfig = {
    id:string;
    baseUrl:string;
    schedule?:TaskScheduleDefinition;
};

export const TypeKnativeEvent = 'eventType';
export const TypeKnativeBroker = 'broker';
export const SystemKnative = 'knative-event-mesh';
export const OwnerKnative = 'knative';
