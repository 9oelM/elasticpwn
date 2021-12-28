# Go does not support building binary from non-main package. This is a little script to build standalone binary. 
mkdir ./tmp-binary-build

cp *.go ./tmp-binary-build

cd tmp-binary-build

go mod init github.com/9oelM/elasticpwn/src/tmp-binary-build/lookup-addrs

go mod edit -replace github.com/9oelM/elasticpwn/core/util=../../util 

go mod tidy

# Replace all occurences 
sed -i 's/package EPLookup_addrs/package main/g' *.go

go build -v

cp lookup-addrs ../

cd ..

rm -rf ./tmp-binary-build
