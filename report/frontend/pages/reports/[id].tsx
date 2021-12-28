import * as React from "react"
import { useRouter } from 'next/router'
import { ScanResult } from "../../types/elastic"
import { x } from "@xstyled/styled-components";
import { SF } from "../../styles/fragments";
import { getOnlyNumber, getOnlyString } from "../../util/string";
import PreInfo from "../../templates/localFragments/PreInfo";
import { getUnitScore, SizeUnit } from "../../util/filesize";
import TableInfo from "../../templates/localFragments/Indices";
import Layout from "../../components/layout/layout";
import { GetStaticPaths, GetStaticProps, NextPage } from "next";

import mongoClientPromise, { usefulScanResultFilter } from '../../util/mongo'
import { ObjectId } from "bson";
import { QuickActionButtons } from "../../components/quickActionButtons/quickActionButtons";
import { Config } from "../../config/env";
import { enhance, tcAsync } from "../../util/react-essentials";
import { API } from "../../util/api";
import { useOnReviewedAndOnPass } from "../../hooks/useOnReviewedAndOnPass";
import { usePersistentDatabaseServerAvailableStatus } from "../../hooks/usePersistentDatabaseServerAvailableStatus";
import Link from "next/link";
import { LocalStorageKeys, LocalStorageManager } from "../../util/localStorage";

type ReportProps = {
    id: string
    scanResult: ScanResult,
    nextScanResult: ScanResult | null
}

const Report: NextPage<ReportProps> = enhance<ReportProps>(({
    scanResult: scanResultRaw,
    nextScanResult: nextScanResultRaw,
}) => {
    const { scanResult } = scanResultRaw

    const [wasCurrentRootUrlReviewed, setCurrentRootUrlReviewed] = React.useState(false)
    React.useEffect(() => {
        const allReviewedUrls = LocalStorageManager.getValue(LocalStorageKeys.REVIEWED)

        if (allReviewedUrls === null) {
            return
        }
        setCurrentRootUrlReviewed(Boolean(allReviewedUrls[scanResult.rootUrl]))
    }, [scanResult.rootUrl])

    const indicesSortedBySize = React.useMemo(() => {
        if (!scanResult || !scanResult.indices) return null

        return [...scanResult.indices].sort((a, b) => {
            if (!a["store.size"]) {
                return 1
            } else if (!b["store.size"]) {
                return 0
            }

            const aSizeUnitScore = getUnitScore(getOnlyString(a["store.size"]) as SizeUnit)
            const bSizeUnitScore = getUnitScore(getOnlyString(b["store.size"]) as SizeUnit)

            if (aSizeUnitScore !== bSizeUnitScore) {
                return bSizeUnitScore - aSizeUnitScore
            }

            const aSize = getOnlyNumber(a["store.size"])
            const bSize = getOnlyNumber(b["store.size"])

            if (aSize === null) {
                return 1
            } else if (bSize === null) {
                return 0
            }

            return bSize - aSize
        })
    }, [scanResult])

    const indices = React.useMemo(() => {
        return indicesSortedBySize ? <TableInfo
            headings={[
                `name`,
                `store.size`,
                `docs.count`,
                `docs.deleted`,
                `pri.store.size`,
            ]}
            title="Indices"
        >
            {
                indicesSortedBySize.map((index) => {
                    const isIndexSizeBiggerThanMB =
                        index["store.size"] && getUnitScore(getOnlyString(index["store.size"]) as SizeUnit) > 2 

                    return <x.tr
                        key={index.index}
                        borderWidth={1}
                        borderStyle="solid"
                        borderColor="gray-400"
                    >
                        {/* index.index (name) */}
                        <x.td>{index.index}</x.td>
                        <x.td
                            {...(isIndexSizeBiggerThanMB ? {
                                color: `red-800`
                            } : {})}
                        >{index["store.size"]}</x.td>
                        <x.td>{index["docs.count"]}</x.td>
                        <x.td>{index["docs.deleted"]}</x.td>
                        <x.td>{index["pri.store.size"]}</x.td>
                    </x.tr>
                })
            }
        </TableInfo> : null
    }, [indicesSortedBySize])


    console.log(scanResultRaw)
    console.log(nextScanResultRaw)
    const interestingInfo = React.useMemo(() => {
        if (!scanResult || !scanResult.interestingInfo) return null

        const moreThanTwoDotsInName = scanResult.interestingInfo.moreThanTwoDotsInName || [] 
        const urls = scanResult.interestingInfo.moreThanTwoDotsInName || [] 
        return [
            {
                title: "Contains more than two dots",
                info: [
                    ...new Set([
                        ...moreThanTwoDotsInName,
                        ...urls
                    ])].filter((line) => getOnlyNumber(line) === null).sort()
            },
            {
                title: "Contains '@'",
                info: scanResult.interestingInfo.emails
            },
            {
                title: "Public IPs",
                info: scanResult.interestingInfo.publicIps,
            }
        ].map(({ title, info }) => {
            if (!info || info.length === 0) return null

            return <PreInfo
                key={title}
                {...{
                    title,
                    info: info.join("\n")
                }}
            />
        })
    }, [scanResult])

    const allocations = React.useMemo(() => {
        if (!scanResult) return null

        return scanResult.allocations ? <TableInfo
            headings={[
                "disk.avail",
                "disk.indices",
                "disk.percent",
                "disk.total",
                "disk.used",
                "host",
                "ip",
                "node",
                "shards",
            ]}
            title="Allocations"
        >
            {scanResult.allocations.map((alloc, i) => {
                const isDiskTotalBiggerThanMB = alloc["disk.total"] !== null &&
                    getUnitScore(getOnlyString(alloc["disk.total"]) as SizeUnit) > 2 

                return (
                    <x.tr key={i}>
                        <x.td>{alloc["disk.avail"]}</x.td>
                        <x.td>{alloc["disk.indices"]}</x.td>
                        <x.td>{alloc["disk.percent"]}</x.td>
                        <x.td
                            {...(isDiskTotalBiggerThanMB ? {
                                color: `red-800`
                            }: {})}
                        >{alloc["disk.total"]}</x.td>
                        <x.td>{alloc["disk.used"]}</x.td>
                        <x.td>{alloc["host"]}</x.td>
                        <x.td>{alloc["ip"]}</x.td>
                        <x.td>{alloc["node"]}</x.td>
                        <x.td>{alloc["shards"]}</x.td>
                    </x.tr>
                )
            })}
        </TableInfo> : null
    }, [scanResult])

    const aliases = React.useMemo(() => {
        if (!scanResult) return null

        if (!scanResult.aliases || scanResult.aliases.length === 0) return null

        return  <TableInfo
            headings={[
                "alias",
                "filter",
                "index",
            ]}
            title="Aliases"
        >
            {scanResult.aliases.map((alias, i) => {
                return (
                    <x.tr key={i}>
                        <x.td>{alias["alias"]}</x.td>
                        <x.td>{alias["filter"]}</x.td>
                        <x.td>{alias["index"]}</x.td>
                    </x.tr>
                )
            })}
        </TableInfo>
    }, [scanResult])

    const rawIndicesInfo = React.useMemo(() => {
        if (!scanResult || !scanResult.indicesInfoInJson) return null

        let prettyJSON: string | null = null
        try {
            prettyJSON = JSON.stringify(scanResult.indicesInfoInJson, null, 2)
        } catch (e) {
            console.log(e)
        }

        if (!prettyJSON || prettyJSON?.trim() === '{}') return null

        return <PreInfo
            title="Raw indices information"
            info={prettyJSON}
            height='big'
        />
    }, [scanResult])

    const {
        onReviewed,
        onPass
    } = useOnReviewedAndOnPass(scanResult, nextScanResultRaw)

    if (!scanResult) {
        return <x.h1>scan result is null</x.h1>
    }

    return (
        <Layout>
            <x.main
                {...SF.fullWH}
                {...SF.flex}
                flexDirection="column"
                padding="4"
                bg={"gray-800"}
                overflowY="scroll"
            >
                <Link href="/">
                <x.button
                    {...SF.standardButton}
                    minW="200px"
                    minH="30px"
                    maxH="30px"
                >
                ← back to main
                </x.button>
                    </Link>
                <x.h1
                    fontSize={{ md: "6xl", xs: "4xl" }}
                    color="gray-400"
                    fontWeight="bold"
                >{`Report: `}
                    <x.a
                        color="gray-400"
                        href={`http://${scanResult.rootUrl}`}
                        target="_blank"
                        rel="noreferrer noopener"
                        style={SF.wrapLines}
                        textDecoration="underline"
                    >
                        {scanResult.rootUrl}
                    </x.a>
                </x.h1>
                {wasCurrentRootUrlReviewed ? <x.p
                    color="red-300"
                    pt={1}
                    pb={1}
                >
                    This URL has been reviewed once by you.
                </x.p> : null}
                <x.section
                    {...SF.fullWH}
                    {...SF.flex}
                    flexDirection="column"
                    bg={"transparent"}
                    spaceY={4}
                >
                    {indices}
                    {scanResult.interestingWords ? <PreInfo
                        title="Word-based interesting regex matches"
                        info={scanResult.interestingWords.join("\n")}
                    /> : null}
                    {interestingInfo}
                    {allocations}
                    {aliases}
                    {rawIndicesInfo}
                    <x.div minH={{ xs: '50px', md: `150px` }}></x.div>
                </x.section>
            </x.main>
            <QuickActionButtons 
                onReviewed={onReviewed}
                onPass={onPass}
            />
        </Layout>
    )
})()

export default Report

export const getStaticProps: GetStaticProps<{}, { id: string }> = async (context) => {
    if (!context.params) {
        throw new Error(`context.params is undefined or null`)
    }

    console.log(`context.params.id: ${context.params.id}`)

    const client = await mongoClientPromise
    const elasticpwnDb = await client.db(Config.DB_NAME)
    const collection = elasticpwnDb.collection<ScanResult>(Config.COLLECTION_NAME)

    const currentAndNextScanResultsCursor = await collection.find({
        $and: [
            {
                _id: {
                    /**
                     * the documents are just sorted by id (insertion order)
                     */
                    $gte: new ObjectId(context.params.id)
                },
            },
            usefulScanResultFilter,
        ]
    }).limit(2)

    const singleScanResult = await currentAndNextScanResultsCursor.next()
    const nextSingleScanResult = (await currentAndNextScanResultsCursor.hasNext()) ? await currentAndNextScanResultsCursor.next() : null

    // https://github.com/vercel/next.js/issues/11993#issuecomment-617375501
    const scanResult = JSON.parse(JSON.stringify(singleScanResult))
    const nextScanResult = JSON.parse(JSON.stringify(nextSingleScanResult))
    
    return {
        props: {
            // to enable destructuring in the component
            scanResult: scanResult === null ? { scanResult: null } : scanResult,
            nextScanResult: nextScanResult,
        }
    }
}

export const getStaticPaths: GetStaticPaths<{ id: string }> = async () => {
    const client = await mongoClientPromise
    const elasticpwnDb = await client.db(Config.DB_NAME)
    const collection = elasticpwnDb.collection<ScanResult>(Config.COLLECTION_NAME)

    const allInterestingScanResultsIds =
       await collection
        .find(usefulScanResultFilter)
        .map(({ _id }) => ({ id: _id.toString() })).toArray()

    return {
        paths: allInterestingScanResultsIds.map(({ id }) => ({ params: { id }})), // all pages with interesting scan result need to be created at build time
        // this needs to be set as false to enable `next export` command
        fallback: false
    }
}