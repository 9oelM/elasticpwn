import { ObjectId } from "bson";

export interface ElasticProductInfo {
    _id: string
    rootUrl: string
    // {"index":"index_name","docs.count":"2355","docs.deleted":"0","store.size":"3.8mb","pri.store.size":"3.8mb"}
    indices: null | {
        index: string
        "docs.count": string
        "docs.deleted": string
        "store.size": string
        "pri.store.size": string
    }[]
    // this data is too complex. It will be presented in <pre> tag, so just render it as string
    indicesInfoInJson: string
    interestingWords: string[]
    interestingInfo: null | {
        emails: string[]
        urls: string[]
        publicIps: string[]
        moreThanTwoDotsInName: string[]
    }
    // {"alias":".kibana","filter":"-","index":".kibana_1","is_write_index":"-","routing.index":"-","routing.search":"-"}
    aliases: null | {
        alias: string
        filter: string
        index: string
    }[]
    // {"disk.avail":"133.8gb","disk.indices":"83.2mb","disk.percent":"13","disk.total":"154.8gb","disk.used":"21gb","host":"10.42.197.128","ip":"10.42.197.128","node":"fde99ab8806e","shards":"13"}
    allocations: null | {
        "disk.avail": string | null
        "disk.indices": string | null
        "disk.percent": string | null
        "disk.total": string | null
        "disk.used": string | null
        "host": string | null 
        "ip": string | null 
        "node": string | null 
        "shards": string | null
    }[]
    isInitialized: boolean

    // unused properties (for now)

    hasAtLeastOneIndexSizeOverGB: boolean
    // {"action":"cluster:monitor/nodes/stats","ip":"10.42.197.128","node":"fde99ab8806e","parent_task_id":"-","running_time":"3.4ms","start_time":"1635165293843","task_id":"_-IU2jWcQsaT0XsJCga1mg:8439865","timestamp":"12:34:53","type":"transport"}
    tasks: Record<string, string>[]
    created_at: string
    trainedModels: string
    transforms: string
    // not really interested in these anyway
    count: null
    master: null
    nodeaatrs: null
    nodes: null
    pendingTasks: null
    plugins: null
}

export interface ScanResult {
    _id: ObjectId
    scanResult: ElasticProductInfo
}