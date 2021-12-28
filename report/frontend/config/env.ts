export const Config = {
    DB_NAME: process.env.DB_NAME as string,
    COLLECTION_NAME: process.env.COLLECTION_NAME as string,
    SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH: process.env.SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH as string | undefined
}

Object.keys(Config).forEach((key) => {
    console.log(`Found env key: ${key}`)
    if (key === `SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH`) return

    if (key === undefined || key === null) {
        throw new Error(`Config['${key}'] is undefined or null. Please check .env.local file.`)
    }
})