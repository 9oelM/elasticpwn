#!/bin/bash
echo "
############################################

before running this script, 
1. you should have configured .env.local in /frontend properly
2. you should have a running mongodb server containing scan results

############################################
"

cd frontend

original_nextjs_config=$(cat next.config.js)

function cleanup {
  echo "${original_nextjs_config}" > next.config.js
}

trap cleanup EXIT

# https://unix.stackexchange.com/questions/356312/sed-insert-something-after-the-second-last-line
sed -i '$i  ,assetPrefix: ".",' next.config.js

npm run build

npm run "export"

echo "${original_nextjs_config}" > next.config.js
