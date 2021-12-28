import { x } from '@xstyled/styled-components'
import { Filter } from 'mongodb'
import type { GetStaticProps, NextPage } from 'next'
import Link from 'next/link'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import Layout from '../components/layout/layout'
import { Config } from '../config/env'
import { usePersistentDatabaseServerAvailableStatus } from '../hooks/usePersistentDatabaseServerAvailableStatus'
import { SF } from '../styles/fragments'
import TableInfo from '../templates/localFragments/Indices'
import { ElasticProductInfo, ScanResult } from '../types/elastic'
import { API, APIStatus } from '../util/api'
import { LocalStorageInternal, LocalStorageKeys, LocalStorageManager } from '../util/localStorage'
import mongoClientPromise, { usefulScanResultFilter } from '../util/mongo'
import { tcAsync } from '../util/react-essentials'

const Home: NextPage<{
  count: number,
  allInterestingScanResults: { rootUrl: string; id: string; }[]
  allUninterestingScanResultsUrls: { rootUrl: string }[]
}> = ({
  count,
  allInterestingScanResults,
  allUninterestingScanResultsUrls,
}) => {
    const [
      newInterstingScanResults,
      setNewInterestingScanResults
    ] = useState<{ rootUrl: string; id: string; }[]>([])
    const [
      newUninterestingUrls,
      setNewUninterestingUrls
    ] = useState<ScanResult['scanResult']['rootUrl'][]>([])

    const [
      reviewedUrlsFromPersistentDatabase,
      setReviewedUrlsFromPersistentDatabase
    ] = useState<{ rootUrl: string; id: string; }[]>([])

    const [reviewedUrlsInCurrentSession, setReviewedUrlsInCurrentSession] = useState<null | LocalStorageInternal[LocalStorageKeys.REVIEWED]>(null)
    // https://github.com/vercel/next.js/discussions/19911
    useEffect(() => {
      const reviewedUrls = LocalStorageManager.getValue(LocalStorageKeys.REVIEWED)
      setReviewedUrlsInCurrentSession(reviewedUrls)
    }, [])

    const persistentDatbaseServerAvailableStatus = usePersistentDatabaseServerAvailableStatus()

    useEffect(() => {
      async function processUrls() {
        const [err, reviewedUrlsFromPersistentDatabase] = await tcAsync(API.getReviewedUrls())

        if (err || !reviewedUrlsFromPersistentDatabase || !reviewedUrlsFromPersistentDatabase.urls) {
          console.error(`could not get reviewed urls`)
          setNewUninterestingUrls(allUninterestingScanResultsUrls.map(({ rootUrl }) => rootUrl))
          return
        }

        // @ts-ignore
        setReviewedUrlsFromPersistentDatabase(reviewedUrlsFromPersistentDatabase.urls)
        setNewInterestingScanResults(
          allInterestingScanResults.filter(({ rootUrl }) => {
            return !reviewedUrlsFromPersistentDatabase.urls?.find((url) => rootUrl === url)
          })
        )
        // for the case where you've previously reviewed the same urls already
        setNewUninterestingUrls(
          allUninterestingScanResultsUrls.filter(({ rootUrl }) => {
            return !reviewedUrlsFromPersistentDatabase.urls?.find((url) => rootUrl === url)
          }).map(({ rootUrl }) => rootUrl)
        )
      }

      if (persistentDatbaseServerAvailableStatus !== APIStatus.LOADING) {
        processUrls()
      }
    }, [allInterestingScanResults, allUninterestingScanResultsUrls, persistentDatbaseServerAvailableStatus])

    const onSaveUninterestingUrlsAsReviewed = useCallback(async () => {
      const [err, result] = await tcAsync(API.postUrlsAsReviewed(newUninterestingUrls))
      if (err || result?.error !== null) {
        console.error(`failed to save uninteresting urls`)
        return
      }
    }, [newUninterestingUrls])

    const uninterestingUrlsInstruction = React.useMemo(() => {
      return newUninterestingUrls.length > 0 ?
        <x.section>
          <x.p color='gray-400' fontSize={{ md: `1g`, xs: `base` }}>
            If you are using a mongodb backend to store your reviewed data, you should click the button below to save all uninteresting urls (total of {newUninterestingUrls.length}) collected for this report as reviewed in your mongodb database.
      </x.p>
          <x.p color='gray-400' fontSize={{ md: `1g`, xs: `base` }}>
            If you don&lsquo;t, you will face these urls again in your report for the next recon (if you are running the scan with the corresponding option).
        {` `}
            <x.span color='red-300'>
              If you think something went wrong while scanning, re-run the scanning and generate the report again instead of clicking the button, because the change is irreversible.
        </x.span>
          </x.p>
          <x.button
            {...SF.standardButton}
            w={150}
            mt={4}
            mb={4}
            bg={{ hover: 'red-800', _: `red-900` }}
            color='gray-300'
            onClick={onSaveUninterestingUrlsAsReviewed}
          >
            Save all {newUninterestingUrls.length} uninteresing URLs as reviewed
      </x.button>
        </x.section> : null
    }, [newUninterestingUrls, onSaveUninterestingUrlsAsReviewed])

    return (
      <Layout>
        <x.main {...SF.fullWH} p={{ md: 10, xs: 5 }}>
          <x.header spaceY={2} pb={2} >
            <x.h1 color='gray-400' fontSize={{ md: `3xl`, xs: `2xl` }}>elasticpwn report</x.h1>
            {
              persistentDatbaseServerAvailableStatus === APIStatus.SUCCESSFUL ?
                <TableInfo
                  title=""
                  headings={["Total", "To be reviewed", "reviewed", "uninteresting"]}
                >
                  <x.tr>
                    <x.td textAlign='center'>{count}</x.td>
                    <x.td textAlign='center'>{newInterstingScanResults.length}</x.td>
                    <x.td textAlign='center'>{reviewedUrlsFromPersistentDatabase.length}</x.td>
                    <x.td textAlign='center'>{newUninterestingUrls.length}</x.td>
                  </x.tr>
                </TableInfo> : <TableInfo
                  title=""
                  headings={["Total", "pages generated", "uninteresting (no pages generated for these)"]}
                >
                  <x.tr>
                    <x.td textAlign='center'>{count}</x.td>
                    <x.td textAlign='center'>{newInterstingScanResults.length}</x.td>
                    <x.td textAlign='center'>{newUninterestingUrls.length}</x.td>
                  </x.tr>
                </TableInfo>
            }

          </x.header>
          {persistentDatbaseServerAvailableStatus === APIStatus.SUCCESSFUL ? uninterestingUrlsInstruction : null}
          <x.section
            {...SF.flex}
            w="100%"
            justifyContent="space-between"
          >
            <x.article>
              <x.h2 color="gray-400" pb={2}>
                Reviewed in current session (based on localStorage data):
              </x.h2>
              <x.ul spaceY={1}>
                {allInterestingScanResults.map(({ rootUrl, id }) => {
                  if (reviewedUrlsInCurrentSession?.[rootUrl]) {
                    return (
                      <Link href={`/reports/${id}?rootUrl=${rootUrl}`} key={id} passHref>
                        <x.li fontSize='xl' color='gray-400' textDecoration='underline' bg={{ _: 'none', 'hover': 'gray-700' }} style={SF.cursorPointer}>
                          {rootUrl}
                        </x.li>
                      </Link>
                    )
                  }
                  return null
                })
                }
              </x.ul>
            </x.article>
            <x.article>
              <x.h2 color="gray-400" pb={2}>
                To be reviewed in current session (based on localStorage data):
              </x.h2>
              <x.ul spaceY={1}>
                {(persistentDatbaseServerAvailableStatus === APIStatus.SUCCESSFUL ? newInterstingScanResults : allInterestingScanResults).map(({ rootUrl, id }) => {
                  if (reviewedUrlsInCurrentSession?.[rootUrl]) {
                    return null
                  }
                  return (
                    <Link href={`/reports/${id}?rootUrl=${rootUrl}`} key={id} passHref>
                      <x.li fontSize='xl' color='gray-400' textDecoration='underline' bg={{ _: 'none', 'hover': 'gray-700' }} style={SF.cursorPointer}>
                        {rootUrl}
                      </x.li>
                    </Link>
                  )
                })
                }
              </x.ul>
            </x.article>
          </x.section>
        </x.main>
      </Layout>
    )
  }

export default Home

export const getStaticProps: GetStaticProps = async (context) => {
  const client = await mongoClientPromise
  const elasticpwnDb = await client.db(Config.DB_NAME)
  const collection = elasticpwnDb.collection<ScanResult>(Config.COLLECTION_NAME)
  const count = await collection.count()
  /**
   * @todo cache results in development
   */
  const allInterestingScanResults =
    await collection
      .find(usefulScanResultFilter)
      .map(({
        scanResult,
        // for findById in child pages
        _id
      }) => ({
        rootUrl: scanResult.rootUrl,
        id: _id.toString()
      }))
      .toArray()
  const allUninterestingScanResultsUrls =
    await collection
      .find({ $nor: [usefulScanResultFilter] })
      .map(({
        scanResult,
      }) => ({
        rootUrl: scanResult.rootUrl,
      }))
      .toArray()

  return {
    props: {
      count,
      allInterestingScanResults,
      allUninterestingScanResultsUrls,
    },
    // revalidate every 30 mins to speed up build speed while developing
    revalidate: 60 * 30
  }
}