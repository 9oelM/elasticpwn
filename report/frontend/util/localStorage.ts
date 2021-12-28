import { ScanResult } from "../types/elastic"

export enum LocalStorageKeys { 
    NOT_REVIEWED = `elasticpwn_NOT_REVIEWED`,
    REVIEWED = `elasticpwn_REVIEWED`,
    COME_BACK_LATER = `elasticpwn_COME_BACK_LATER`, 
}

export interface LocalStorageInternal {
    [LocalStorageKeys.NOT_REVIEWED]: Record<ReturnType<ScanResult['_id']['id']['toString']>, boolean>
    [LocalStorageKeys.REVIEWED]: Record<ReturnType<ScanResult['_id']['id']['toString']>, boolean>
    [LocalStorageKeys.COME_BACK_LATER]: Record<ReturnType<ScanResult['_id']['id']['toString']>, boolean>
}

export class LocalStorageManager {
    static getValue<T extends keyof LocalStorageInternal>(key: LocalStorageKeys): LocalStorageInternal[T] | null {
        const item = localStorage.getItem(key)
        
        if (item === null) return null

        let parsed: null | any = null

        try {
            parsed = JSON.parse(item)
        } catch {
            return null
        }

        return parsed
    }

    static setValue<Key extends keyof LocalStorageInternal, Value extends LocalStorageInternal[Key]>(key: LocalStorageKeys, value: Value): boolean {
        let valueStr: string | null = null 
        
        try {
            valueStr = JSON.stringify(value)
        } catch {
            return false
        }

        localStorage.setItem(key, valueStr)

        return true
    }

    static addStringToDict<Key extends keyof LocalStorageInternal>(localStorageKey: LocalStorageKeys, key: string): boolean {
        const existingValue = LocalStorageManager.getValue(localStorageKey) ?? {}

        existingValue[key] = true

        return LocalStorageManager.setValue(localStorageKey, existingValue)
    }
}