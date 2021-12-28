import { useState, useEffect } from "react"
import { APIStatus, API } from "../util/api"
import { tcAsync } from "../util/react-essentials"

// just call this everywhere, without using a global state
// to just make things very simple
export function usePersistentDatabaseServerAvailableStatus() {
  const [
    persistentDatbaseServerAvailableStatus,
    setPersistentDatbaseServerAvailable
  ] = useState<APIStatus>(APIStatus.LOADING)

  useEffect(() => {
    async function checkPersistentDatabaseServerAvailable() {
      const [err] = await tcAsync(API.ping())
      setPersistentDatbaseServerAvailable(err === null ? APIStatus.SUCCESSFUL : APIStatus.FAILED)
    }
    checkPersistentDatabaseServerAvailable()
  }, [])

  return persistentDatbaseServerAvailableStatus
}