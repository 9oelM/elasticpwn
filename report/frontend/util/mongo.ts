import { Filter, MongoClient, MongoClientOptions } from 'mongodb'
import { ScanResult } from '../types/elastic'

const uri = process.env.MONGODB_URI as string

let client
let clientPromise: Promise<MongoClient>

if (!process.env.MONGODB_URI) {
  throw new Error('Please add your Mongo URI to .env.local')
}

if (process.env.NODE_ENV === 'development') {
  // In development mode, use a global variable so that the value
  // is preserved across module reloads caused by HMR (Hot Module Replacement).
  // @ts-ignore
  if (!global!._mongoClientPromise) {
    client = new MongoClient(uri)
    // @ts-ignore
    global._mongoClientPromise = client.connect()
  }
  // @ts-ignore
  clientPromise = global._mongoClientPromise
} else {
  // In production mode, it's best to not use a global variable.
  client = new MongoClient(uri)
  clientPromise = client.connect()
}

// Export a module-scoped MongoClient promise. By doing this in a
// separate module, the client can be shared across functions.
export default clientPromise

export const usefulScanResultFilter: Filter<ScanResult> = {
    $or: [{
      'scanResult.aliases.0': { $exists: true }
    }, {
      'scanResult.allocations.0': { $exists: true }
    }, {
      'scanResult.indices.0': { $exists: true }
    }, {
      'scanResult.indicesInfoInJson': { $ne: null }
    }]
}
