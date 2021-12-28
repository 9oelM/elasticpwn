/** @type {import('next').NextConfig} */
module.exports = {
  reactStrictMode: true,
  distDir: 'build',
  // https://stackoverflow.com/questions/66137368/next-js-environment-variables-are-undefined-next-js-10-0-5
  env: {
    SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH: process.env.SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH
  }
  // https://github.com/vercel/next.js/discussions/13578
  // this needs to be uncommented when npm run build
  // assetPrefix: '.',
}
