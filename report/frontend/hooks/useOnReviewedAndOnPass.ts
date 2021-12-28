import { useRouter } from "next/router";
import React from "react";
import { ScanResult } from "../types/elastic";
import { API, APIStatus } from "../util/api";
import { LocalStorageKeys, LocalStorageManager } from "../util/localStorage";
import { tcAsync } from "../util/react-essentials";

export function useOnReviewedAndOnPass(
    scanResult: ScanResult['scanResult'],
    nextScanResultRaw: ScanResult | null,
    ) {

    const router = useRouter()
    const browseToNextReport = React.useCallback(() => {
        if (!nextScanResultRaw) return;

        router.push(`/reports/${nextScanResultRaw._id}?rootUrl=${nextScanResultRaw.scanResult.rootUrl}`)
    }, [nextScanResultRaw])
    const onReviewed = React.useCallback(async () => {
        const [err, postUrlsAsReviewedResult] = await tcAsync(API.postUrlsAsReviewed([scanResult.rootUrl])) 
        LocalStorageManager.addStringToDict(LocalStorageKeys.REVIEWED, scanResult.rootUrl)
        
        browseToNextReport()
        if (err) {
            console.log(err)
            return;
        }
        const hasAnotherError = postUrlsAsReviewedResult?.error !== null
        if (hasAnotherError) {
            console.log(postUrlsAsReviewedResult)
            return;
        }
    }, [browseToNextReport, scanResult.rootUrl])
    const onPass = React.useCallback(() => {
        browseToNextReport()
    }, [browseToNextReport])

    React.useEffect(() => {
        const listener = (ev: KeyboardEvent) => {
            if (!ev.altKey) return

            switch (ev.key) {
                case `r`: {
                    onReviewed()
                    break
                }
                case `c`: {
                    onPass()
                    break
                }
            }
        }
        document.addEventListener(`keydown`, listener)

        return () => {
            document.removeEventListener(`keydown`, listener)
        }
    }, [onReviewed, onPass])

    return {
        onReviewed,
        onPass
    }
}