import {TaskScheduleDefinition} from '@backstage/backend-tasks';

export type KnativeEventMeshProviderConfig = {
    id:string;
    baseUrl:string;
    schedule?:TaskScheduleDefinition;
};
