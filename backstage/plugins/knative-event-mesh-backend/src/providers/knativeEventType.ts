import {ApiEntity} from "@backstage/catalog-model";

export interface KnativeEventType extends ApiEntity {
    consumedBy?:string[];
}
