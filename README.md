# elasticpwn

Quickly collects data from exposed Elasticsearch or Kibana instances and generates a report to be reviewed. It mainly aims for sensitive data exposure and content discovery. 

# Table of contents
* [elasticpwn](#elasticpwn)
* [Table of contents](#table-of-contents)
* [Rationale](#rationale)
* [Collecting data](#collecting-data)
* [Generating a report](#generating-a-report)
* [Generating a report with a persistent MongoDB database](#generating-a-report-with-a-persistent-mongodb-database)
* [Generating a report without a persistent MongoDB database](#generating-a-report-without-a-persistent-mongodb-database)
* [Notes about performance](#notes-about-performance)
   * [Threads (-t option)](#threads--t-option)
   * [Maximum number of indices to request (-max-i option)](#maximum-number-of-indices-to-request--max-i-option)
   * [Maximum size of an index to request (-max-is option)](#maximum-size-of-an-index-to-request--max-is-option)
   * [Generating a report with many pages](#generating-a-report-with-many-pages)
* [Full CLI reference](#full-cli-reference)
   * [elasticpwn](#elasticpwn-1)
   * [elasticpwn-backend](#elasticpwn-backend)
* [Contributing](#contributing)

# Rationale

Why do you need this project? If you are hunting for sensitive data exposure or content discovery, Elasticsearch and Kibana could be your target, because they do not have an authentication turned on by default and there are thousands of exposed instances on the Internet. But as you try to look into the data acquired from these instances, you need a lot of ad-hoc data processing and manual work to find out sensitive pieces of data. To be able to easily do that, we need some neat way to collect data, process it, and review them all at once. This is what `elasticpwn` tries to achieve.

This project is different from other similar OSS projects in that they actually request detailed indices info by using **[`_search` APIs](https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html).** This greatly reduces the burden of visiting each instance manually and querying the data of an index that looks interesting. Also, it automatically ignores data that looks quite useless, because many open Elasticsearch and Kibana instances may have sample/boilerplate data which are just useless to inspect. The list of this should be constantly updated. [More details in CONTRIBUTING.md](CONTRIBUTING.md). 

# Collecting data
1. You need `go^1.17` to install `elasticpwn`.
1. Install `elasticpwn`:
      ```bash
      go install github.com/9oelM/elasticpwn/elasticpwn@latest
      ```
1. You will need to have a list of URLs to try against. Get them from OSINT platforms like [shodan.io](https://shodan.io) or [binaryedge.io](https://binaryedge.io). Download them and put them into a text file, with each ip and port written as follows:
      ```
      123.123.123.123:80
      124.124.124.124:443
      ... and so on.
      ```
1. Think about which option to choose for storing output. If you are collecting a really large sum of data, from say, 2000 instances of kibana, then you would probably need to use `mongo` instead of `json`, because `mongo` is the only option where you could generate a report. Otherwise, choose `json` or `plain`. If you will use `mongo`, preferrably launch a local mongodb instance. You can easily do it by using docker-compose file provided at the root of this repository: `curl https://raw.githubusercontent.com/9oelM/elasticpwn/main/docker-compose-mongo-only.yml -o docker-compose-mongo-only.yml && docker-compose -f docker-compose-mongo-only.yml up -d`
1. Run `elasticpwn` and wait for the data collection to be finished. For example:
      ```
      elasticpwn elasticsearch -f list_of_elasticsearch_instances.txt -murl mongodb://root:example@172.17.0.1:27017/ -of mongo -t 12     
      ```
      if you are not going for mongo, then it would be something like:
      ```
      elasticpwn elasticsearch -f list_of_elasticsearch_instances.txt -of json -t 12 
      ```
1. After it is finished, check data is properly collected.

# Generating a report

**Generating a report assumes pre-installation of `npm` and `node`**.

Reading the report is a convenient way to review data obtained from the search done by the CLI. The report contains data not limited to:
- URL of the elasticsearch/kibana instance
- indices
- detailed info about each index (POST `/<index_name>/_search` result)
- interesting information (may be relevant to sensitive information disclosure)
- nodes
- allocations
- aliases

Currently, elasticsearch and kibana plugins do not collect any more data than this despite the ability to do so, because other data do not usually include sensitive information. If you think this is not the case, please open an issue.

Due to performance reasons, generating an report is only possible by querying data from mongodb. **JSON backend is not supported.**

This is a little preview of how a report will look like:

![report root](./screenshots/REPORT_ROOT_SCREENSHOT.png)
![report](./screenshots/REPORT_SCREENSHOT.png)

# Generating a report with a persistent MongoDB database
`elasticpwn` offers a way to store the history of the reviewed IPs in a database, so that you will know you don't have to review them again in the future if you encounter them. This is different from the database that data collected from elasticsearch/kibana instances are put into. 

1. Generate the report. This will create a directory called `report` under CWD:
      ```bash
      # -cn should be either elasticsearch or kibana
      # -murl should be where the collected data is stored at
      # -dn should be the url where elasticpwn-backend will be hosted
      elasticpwn report generate -cn elasticsearch -murl mongodb://root:example@172.17.0.1:27017/ -dn http://localhost:9292
      ```
1. To connect to a persistent MongoDB database (you will store the list of all IPs you reviewed), install `elasticpwn-backend`: 
      ```
      go install github.com/9oelM/elasticpwn/report/elasticpwn-backend@latest
      ```
1. Set up a persistent MongoDB database. It could be your local datbase, or a remote one. I would highly recommend signing up for a free account for [MongoDB Atlas](https://www.mongodb.com/cloud/atlas/register) because it offers 500MB for free, and there is no chance that you will reach maximum usage just by putting the list of all IPs you reviewed (note that it won't store data from the indices and so on - only IPs are stored).
1. Get the mongodb URI from MongoDB Atlas or somewhere else you managed to use, for example: `mongodb+srv://my-mongo:mypassword@my-mongodb-name.tn3al.mongodb.net/defaultcollectionname?retryWrites=true&w=majority`.  
1. Launch server that connects with MongoDB URI above. You can also provide values through `.env.local` file. See [`/report/backend/.env.example`](./report/backend/.env.example). Below is an example command that serves reviewed urls of elasticsearch:
      ```bash
      elasticpwn-backend -mongodbUri=mongodb+srv://username:pw@somewhere.mongodb.net/default-collection-name?retryWrites=true&w=majority -databaseCollectionName=elasticsearch_reviewed_urls -databaseName elasticpwn -port 9292
      ```
1. If everything is successfully bootstrapped, you should see some logs coming up:
      ```bash
      2021/12/28 05:38:54 Connected to database
      2021/12/28 05:38:54 Inserting
      2021/12/28 05:38:54 Test url already created before
      [GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

      [GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
      - using env:   export GIN_MODE=release
      - using code:  gin.SetMode(gin.ReleaseMode)

      [GIN-debug] GET    /ping                     --> main.PingHandler (4 handlers)
      [GIN-debug] GET    /urls                     --> main.GetUrlsHandlerGenerator.func1 (4 handlers)
      [GIN-debug] POST   /urls                     --> main.PostUrlsHandlerGenerator.func1 (4 handlers)
      [GIN-debug] DELETE /urls                     --> main.DeleteUrlsHandlerGenerator.func1 (4 handlers)
      [GIN-debug] Listening and serving HTTP on 0.0.0.0:9292
      ```
      Run `curl http://localhost:9292/urls` or `curl http://localhost:9292/ping` to check if it's running correctly. 
1. With this server bootstrapped and running, view the report at the same time:
      ```
      elasticpwn report view -d ./path-to-report -p 9999
      ```
      Then, you will be able to see the report at http://localhost:9999.

# Generating a report without a persistent MongoDB database
If you are just running `elasticpwn` for once or twice, generating a report without a persistent database may be the right choice.

1. Generate the report by running the command below (After collecting data by running `elasticpwn elasticsearch|kibana`). This will create a directory called `report` under CWD:
      ```bash
      # -cn should be either elasticsearch or kibana
      # -murl should be where the collected data is stored at
      elasticpwn report generate -cn elasticsearch -murl mongodb://root:example@172.17.0.1:27017/
      ```

1. View the report:
      ```bash
      elasticpwn report view -d ./path-to-report -p 9999
      ```
      Then, you will be able to see the report at http://localhost:9999.

# Notes about performance
## Threads (`-t` option)
`elasticpwn` goes through extensive regex matching work to find interesting words that may be relevant to sensitive information disclosure. Therefore it is recommended to keep the number of threads at about the number of your computer's cores (output from the command `nproc`). Otherwise, the program may crash or slow down. 

For example, if the number of cores of your computer is 12, keep the number of threads at about 12~24.

## Maximum number of indices to request (`-max-i` option)
If this is too large, it might cause MongoDB to reject insertion of data due to its size. Stick with the default option if you are unsure. 

## Maximum size of an index to request (`-max-is` option)
If this is too large, it might cause MongoDB to reject insertion of data due to its size. Stick with the default option if you are unsure. This can also affect RAM and CPU usage.

## Generating a report with many pages
If you are collecting data from.. say, 5000 instances, generating a report with that many pages with _javascript_ could be a bit of challenge. Next.js is used to create it (Gatsby.js failed because it could not hold this much data and would just fail due to memory shortage). If you are collecting very much of data, `elasticpwn` could expect 4GB to 8GB of vacant RAM. 

# Full CLI reference

## `elasticpwn`
```
Usage: elasticpwn [elasticsearch|kibana [...plugin options] or elasticpwn report [generate|view] [...plugin options]
[elasticsearch] plugin options:
  -f string
        [REQUIRED] path to a file with urls (url per line)
  -max-i int
        maximum number of indices to request. 
        If you intend to set this as a high number, make sure you've got enough storage. 
        If you don't know what an index is, 
        refer to elasticserach docs at https://www.elastic.co/blog/what-is-an-elasticsearch-index (default 5)
  -max-is int
        maximum size of an index to request. 
        If you intend to set this as a high number, make sure you've got enough storage.
        If you don't know what 'size' is,
        please refer to elasticsearch docs on
        '<endpoint>/_cat/_search?size=' at 
        https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html (default 70)
  -murl string
        [OPTIONAL] needed only when -o=mongo is selected. 
        mongodb url with username and pw included.
        Note that 172.17.0.1 is usually default docker host IP. You may want to change this.
         (default "mongodb://root:example@172.17.0.1:27017/")
  -of string
        [OPTIONAL] output file name. Ignored when mongo option is chosen. (default "elasticsearch.json")
  -om string
        [OPTIONAL] output mode. json|mongo|plain. 
        mongo option will require a mongo server to be up.
        plain mode will output json-like object to each line finishing with a comma. 
        For mongo, Local docker mongo instance is recommneded. (check docker-compose.yml and docs) (default "json")
  -t int
        [OPTIONAL] number of threads when running a plugin (default 8)
[kibana] plugin options:
  -f string
        [REQUIRED] path to a file with urls (url per line)
  -max-i int
        maximum number of indices to request. 
        If you intend to set this as a high number, make sure you've got enough storage. 
        If you don't know what an index is, 
        refer to elasticserach docs at 
        https://www.elastic.co/blog/what-is-an-elasticsearch-index (default 5)
  -max-is int
        maximum size of an index to request. 
        If you intend to set this as a high number, make sure you've got enough storage.
        If you don't know what 'size' is, please refer to elasticsearch docs on
        '<endpoint>/_cat/_search?size=' at 
        https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html (default 70)
  -murl string
        [OPTIONAL] needed only when -o=mongo is selected. 
        mongodb url with username and pw included.
        Note that 172.17.0.1 is usually default docker host IP. You may want to change this.
         (default "mongodb://root:example@172.17.0.1:27017/")
  -of string
        [OPTIONAL] output file name. Ignored when mongo option is chosen. (default "kibana.json")
  -om string
        [OPTIONAL] output mode. json|mongo|plain. 
                mongo option will require a mongo server to be up.
                plain mode will output json-like object to each line finishing with a comma. 
        For mongo, local docker mongo instance is recommneded. (check docker-compose.yml and docs) (default "json")
  -t int
        [OPTIONAL] number of threads when running a plugin (default 8)
[report] generate plugin options:
  -cn string
        [REQUIRED] 
        must be elasticsearch|kibana. Collection name of the mongodb database to be used.
  -dn string
        [OPTIONAL] 
        backend url of persistent database server being used. This should be the url where elasticpwn-backend is hosted. (default "http://localhost:9292")
  -murl string
        [OPTIONAL] 
        mongodb url with username and pw included from which gathered data can be accessed. 
        Note that 172.17.0.1 is usually default docker host IP. You may want to change this. (default "mongodb://root:example@172.17.0.1:27017/")
[report] view plugin options:
  -d string
        [OPTIONAL] the directory of generated report (default "./report")
  -p string
        [OPTIONAL] local port to serve report page from (default "9999")
```

## `elasticpwn-backend`

```
Usage of elasticpwn-bakcend:
  -databaseCollectionName string
        mongodb collection name. Should be any one of: elasticsearch_reviewed_urls|elasticsearch_reviewed_urls_dev|kibana_reviewed_urls|kibana_reviewed_urls_dev
  -databaseName string
        mongodb db name
  -mongodbUri string
        mongodb URI. Example: mongodb+srv://username:pw@somewhere.mongodb.net/default-collection-name?retryWrites=true&w=majority
  -port string
        port at which the server will run. (default "9292")
```

# Contributing
See [CONTRIBUTING.md](./CONTRIBUTING.md).
