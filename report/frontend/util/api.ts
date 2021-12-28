import { Config } from "../config/env"
import { tcAsync } from "./react-essentials"

console.log(Config.SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH)

const rootUrl = Config.SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH

export enum APIStatus {
    LOADING = "LOADING",
    FAILED = "FAILED",
    SUCCESSFUL = "SUCCESSFUL"
}

export class API {
    static ERRORS = Object.freeze({
        NO_ROOT_URL: `NO_ROOT_URL`,
        PERSISTENT_SERVER_DOWN: `PERSISTENT_SERVER_DOWN` 
    })

    static async ping(): Promise<void> {
        const rawResponse = await fetch(`${rootUrl}/ping`);
        rawResponse.json()
    }

    static async postUrlsAsReviewed(urls: string[]): Promise<{ error: null | string }> {
        if (!rootUrl) return { error: API.ERRORS.NO_ROOT_URL }
        const rawResponse = await fetch(`${rootUrl}/urls`, {
            method: 'POST',
            headers: {
              'Accept': 'application/json',
              'Content-Type': 'application/json'
            },
            body: JSON.stringify({ urls })
        });

        const jsonResponse = await rawResponse.json()
        return jsonResponse
    }

    static async getReviewedUrls(): Promise<{ urls?: string[] }> {
        if (!rootUrl) return { urls: undefined }
        const rawResponse = await fetch(`${rootUrl}/urls`, {
            headers: {
              'Accept': 'application/json',
              'Content-Type': 'application/json'
            },
        });

        const jsonResponse = await rawResponse.json()
        return jsonResponse
    }
}