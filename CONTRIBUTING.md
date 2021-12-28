As always, PRs are welcome.

# Spirit
Please bear this spirit in mind when thinking of a fix/new feature proposal/etc: this tool is intended for a user who wants to _analyse_ large OSINT data via semi-automated means. This is the very reason why, unlike other tools, it is designed to request detailed information as much as possible (for example, requesting POST `index_name/_search` for each index). Therefore, please help craft this tool that way when contributing.

# Structure
`/core` contains separate packages, which are direct dependencies of `elasticpwn` package.
The reason for separating different packages is 1) separation of concerns, and 2) there is a possibility that other packages may be used as a standalone module, like `lookup-addrs`.

This is a very simple explanation of what each package does:
- `/elasticpwn`: parsing the CLI flags and running the plugins 
- `/lookup-addrs`: provides utilities for getting info about an URL. This package can be built as a standalone module as well.
- `/util`: contains shared, commonly used tools across packages
- `/plugins`: this is where all plugins reside in. Any new plugins should be built in this folder too. Plugins will be exported and be used in `/elasticpwn`. Plugins should be abstracted well and should only expose as few public methods as possible.

## kibana/elasticsearch plugin
- If you don't know already, elasticsearch is a tool for collecting, analazying and viewing large data, and kibana is the frontend of it.
- Sometimes, an IP would have an elasticsearch but kibana port open. In other occassions, it could be reverse. Otherwise, it could have both open. This is the reason that both plugins are needed.
- Kibana's got dev tools page, and it allows user to send a request to elasticsearch via proxy. So we are using that proxy API.
- Elasticsearch is more straightforward; it has [official API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/rest-apis.html). There is [an official Golang client](https://github.com/elastic/go-elasticsearch) for elasticsearch, but it's not used here because all we need to do for this tool is to query very few API endpoints and it takes not too much effort to do that.
- Elasticsearch can have many useless indices. These are listed in `elastic-util.go` and automatically filtered out when querying indices.

# Code
This is my first project using GoLang. It's very possible that I've made stupid mistakes. Please help fix them.

# How to get started on developing locally

```
$ git clone https://github.com/9oelM/elasticpwn.git

$ cd elasticpwn

$ cd core/main

$ go version # at least 1.17 
go version go1.17.2 linux/amd64

$ go mod download

# alternatively, you can use any other tools that can gather data about it. For example, Google dork.
$ shodan download --limit 300 elasticsearch-sample elasticsearch # you may need to add more keywords to easily find 'open' elasticsearch instances

$ shodan parse --fields ip_str,port --separator : elasticsearch-sample.gz > elasticsearch-sample.txt

$ go build -v

# input file must be a list of URLs pointing to a elasticsearch|kibana instances. For example:
# 123.123.123.123:9200
# 4.5.6.7:9200
# and so on.
$ ./main elasticsearch -f elasticsearch-sample.txt -t 12 -of elasticsearch-sample.json -om json  

# or use mongodb to store result
$ ./main elasticsearch -f elasticsearch-sample.txt -murl mongodb://root:example@172.17.0.1:27017/ -t 12 -om mongo
```

# Debug inside Docker container 
Sometimes you may wanna dive into docker container to inspect mongo directly. You may want to use these commands below in that case.

```
docker container ls # find which container is responsible for mongodb

docker exec -it [container-id] /bin/bash # enter docker container shell

mongo -host "mongodb://root:example@mongo:27017/" -u root -p example # login to console
```
Alternatively, you could use a tool like DataGrip to query from a separate program.

# IDE
Visual Studio Code is highly recommended. [Note that gopls does not support multi-package environment without any config](https://github.com/golang/go/issues/32394). [Please open a workspace on VSCode and add respective package folder (lookup-addrs, main, plugins, ...) to the workspace](https://github.com/golang/go/issues/32394#issuecomment-498385140).

