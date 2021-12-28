// types/globals.d.ts
import type { MongoClient } from 'mongodb'

declare global {
  namespace NodeJS {
    export interface Global {
            _mongoClientPromise: Promise<MongoClient>
        }
    }
}