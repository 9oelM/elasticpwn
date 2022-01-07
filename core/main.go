package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	EPPlugins "github.com/9oelm/elasticpwn/core/plugins"
	EPUtils "github.com/9oelm/elasticpwn/core/util"
)

// elasticsearch and kibana plugins have the exact same fields as of now.
// The reason for the separation of the plugins and the flags is that
// elasticsearch and kibana carry a bit of different traits, so I am a bit afraid
// if anything that's different between them might cause a change in the code in the future.
// therefore I just separated them into two different plugins and flagsets. So please be reminded that this design is by intention.
type Elasticpwn struct {
	mode                 string
	elasticSearchPlugin  *EPPlugins.ElasticSearchPlugin
	kibanaPlugin         *EPPlugins.KibanaPlugin
	reportGeneratePlugin *EPPlugins.ReportGeneratePlugin
	reportViewPlugin     *EPPlugins.ReportViewPlugin
}

var EP_OUTPUT_MODES = []string{"mongo", "json", "plain"}

func printBasicInstruction(
	esPluginFs *flag.FlagSet,
	kibanaPluginFs *flag.FlagSet,
	reportGeneratePluginFs *flag.FlagSet,
	reportViewPluginFs *flag.FlagSet,
) {
	fmt.Println("Usage: elasticpwn [elasticsearch|kibana [...plugin options] or elasticpwn report [generate|view] [...plugin options]")
	fmt.Println("[elasticsearch] plugin options:")
	esPluginFs.PrintDefaults()
	fmt.Println("[kibana] plugin options:")
	kibanaPluginFs.PrintDefaults()
	fmt.Println("[report] generate plugin options:")
	reportGeneratePluginFs.PrintDefaults()
	fmt.Println("[report] view plugin options:")
	reportViewPluginFs.PrintDefaults()
}

func initializeElasticpwn() *Elasticpwn {
	ESPluginFs := flag.NewFlagSet("elasticsearch-plugin", flag.ContinueOnError)
	ESPluginInputFilePath := ESPluginFs.String("f", "", "[REQUIRED] path to a file with urls (url per line)")
	ESPluginThreadsNum := ESPluginFs.Int("t", 8, "[OPTIONAL] number of threads when running a plugin")
	ESPluginOutputMode := ESPluginFs.String("om", "json", `[OPTIONAL] output mode. json|mongo|plain. 
mongo option will require a mongo server to be up.
plain mode will output json-like object to each line finishing with a comma. 
For mongo, Local docker mongo instance is recommneded. (check docker-compose.yml and docs)`)
	ESPluginOutputFilePath := ESPluginFs.String("of", "elasticsearch.json", "[OPTIONAL] output file name. Ignored when mongo option is chosen.")
	ESPluginMongoUrl := ESPluginFs.String("murl", "mongodb://root:example@172.17.0.1:27017/", `[OPTIONAL] needed only when -o=mongo is selected. 
mongodb url with username and pw included.
Note that 172.17.0.1 is usually default docker host IP. You may want to change this.
`)
	// some instances have very many indices, and probably your hard drive will explode if you fetch all
	ESPluginMaxIndicesToCollect := ESPluginFs.Int("max-i", 5, `maximum number of indices to request. 
If you intend to set this as a high number, make sure you've got enough storage. 
If you don't know what an index is, 
refer to elasticserach docs at https://www.elastic.co/blog/what-is-an-elasticsearch-index`)
	// will return 'X' hits from the index
	ESPluginMaxIndexSizeToCollect := ESPluginFs.Int("max-is", 70, `maximum size of an index to request. 
If you intend to set this as a high number, make sure you've got enough storage.
If you don't know what 'size' is,
please refer to elasticsearch docs on
'<endpoint>/_cat/_search?size=' at 
https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html`)

	// kibana plugin has the same options as elasticserach plugin right now,
	// but differences in their behaviors may always make the different, so leave it as it is.
	// as a result, there are many duplicates between es and kibana plugins at least in this file.
	kibanaPluginFs := flag.NewFlagSet("kibana-plugin", flag.ContinueOnError)
	kibanaPluginInputFilePath := kibanaPluginFs.String("f", "", "[REQUIRED] path to a file with urls (url per line)")
	kibanaPluginThreadsNum := kibanaPluginFs.Int("t", 8, "[OPTIONAL] number of threads when running a plugin")
	kibanaPluginOutputMode := kibanaPluginFs.String("om", "json", `[OPTIONAL] output mode. json|mongo|plain. 
	mongo option will require a mongo server to be up.
	plain mode will output json-like object to each line finishing with a comma. 
For mongo, local docker mongo instance is recommneded. (check docker-compose.yml and docs)`)
	kibanaPluginOutputFilePath := kibanaPluginFs.String("of", "kibana.json", "[OPTIONAL] output file name. Ignored when mongo option is chosen.")
	kibanaPluginMongoUrl := kibanaPluginFs.String("murl", "mongodb://root:example@172.17.0.1:27017/", `[OPTIONAL] needed only when -o=mongo is selected. 
mongodb url with username and pw included.
Note that 172.17.0.1 is usually default docker host IP. You may want to change this.
`)
	// some instances have very many indices, and probably your hard drive will explode if you fetch all
	kibanaPluginMaxIndicesToCollect := kibanaPluginFs.Int("max-i", 5, `maximum number of indices to request. 
If you intend to set this as a high number, make sure you've got enough storage. 
If you don't know what an index is, 
refer to elasticserach docs at 
https://www.elastic.co/blog/what-is-an-elasticsearch-index`)
	// will return 'X' hits from the index
	kibanaPluginMaxIndexSizeToCollect := kibanaPluginFs.Int("max-is", 70, `maximum size of an index to request. 
If you intend to set this as a high number, make sure you've got enough storage.
If you don't know what 'size' is, please refer to elasticsearch docs on
'<endpoint>/_cat/_search?size=' at 
https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html`)

	reportGeneratePluginFs := flag.NewFlagSet("report-generate-plugin", flag.ContinueOnError)
	reportGeneratePluginMongoUrl := reportGeneratePluginFs.String("murl", "mongodb://root:example@172.17.0.1:27017/", `[OPTIONAL] 
mongodb url with username and pw included from which gathered data can be accessed. 
Note that 172.17.0.1 is usually default docker host IP. You may want to change this.`)
	reportGeneratePluginCollectionName := reportGeneratePluginFs.String("cn", "", `[REQUIRED] 
must be elasticsearch|kibana. Collection name of the mongodb database to be used.`)
	reportGeneratePluginServerRootUrl := reportGeneratePluginFs.String("dn", "http://localhost:9292", `[OPTIONAL] 
backend url of persistent database server being used. This should be the url where elasticpwn-backend is hosted.`)

	reportViewPluginFs := flag.NewFlagSet("report-view-plugin", flag.ContinueOnError)
	reportViewPluginReportDirectory := reportViewPluginFs.String("d", filepath.FromSlash("./report"), "[OPTIONAL] the directory of generated report")
	reportViewPluginServeAtPort := reportViewPluginFs.String("p", "9999", "[OPTIONAL] local port to serve report page from")

	if len(os.Args) <= 1 {
		fmt.Println("Plugin is not selected. Please try again.")
	}
	if len(os.Args) <= 1 {
		printBasicInstruction(ESPluginFs, kibanaPluginFs, reportGeneratePluginFs, reportViewPluginFs)
		os.Exit(1)
	}
	EPMode := os.Args[1]
	switch EPMode {
	case "elasticsearch":
		{
			if err := ESPluginFs.Parse(os.Args[2:]); err != nil {
				fmt.Println("Wrong options given. Please try again.\nexample: elasticpwn elasticsearch -t=40 -om=json -f=urls.txt")
				os.Exit(1)
			}
		}
	case "kibana":
		{
			if err := kibanaPluginFs.Parse(os.Args[2:]); err != nil {
				fmt.Println("Wrong options given. Please try again.\nexample: elasticpwn kibana -t=40 -om=json -f=urls.txt")
				os.Exit(1)
			}
		}
	case "report":
		{
			if len(os.Args) <= 2 {
				printBasicInstruction(ESPluginFs, kibanaPluginFs, reportGeneratePluginFs, reportViewPluginFs)
				os.Exit(1)
			}
			EPReportMode := os.Args[2]

			switch EPReportMode {
			case "generate":
				{
					if err := reportGeneratePluginFs.Parse(os.Args[3:]); err != nil {
						fmt.Println("Wrong options given. Please try again.\nexample: elasticpwn report generate -murl mongodb://root:example@172.17.0.1:27017/")
						os.Exit(1)
					}
				}
			case "view":
				{
					if err := reportViewPluginFs.Parse(os.Args[3:]); err != nil {
						fmt.Println("Wrong options given. Please try again.\nexample: elasticpwn report view -d ./report")
						os.Exit(1)
					}
				}
			default:
				{
					EPUtils.EPLogger(fmt.Sprintf("\"%v\" is not a valid mode. should be either \"generate\" or \"view\".", EPReportMode))
					os.Exit(1)
				}
			}
		}
	default:
		printBasicInstruction(ESPluginFs, kibanaPluginFs, reportGeneratePluginFs, reportViewPluginFs)
		os.Exit(1)
	}

	switch {
	case ESPluginFs.Parsed():
		{
			needsExit := false

			needsExit = EPUtils.ValidateStringFlag(
				*ESPluginInputFilePath,
				"",
				"-f",
			) ||
				EPUtils.ValidatePositiveInt(
					*ESPluginThreadsNum,
					"-t",
				) ||
				EPUtils.ValidatePositiveInt(
					*ESPluginMaxIndexSizeToCollect,
					"-max-is",
				) ||
				EPUtils.ValidatePositiveInt(
					*ESPluginMaxIndicesToCollect,
					"-max-i",
				)

			switch {
			case EPUtils.Contains(*ESPluginOutputMode, EP_OUTPUT_MODES) == -1:
				{
					fmt.Printf("-om received an invalid value: %v. Possible values: %v\n", *ESPluginOutputMode, strings.Join(EP_OUTPUT_MODES, ","))
					needsExit = true
					break
				}
			case *ESPluginOutputFilePath != "elasticsearch.json" && *ESPluginOutputMode == "mongo":
				{
					EPUtils.EPLogger(fmt.Sprintf("-of option was set as %v but will be ignored because -o option was set as mongo\n", *ESPluginOutputFilePath))
				}
			case *ESPluginOutputMode == "mongo" && *ESPluginMongoUrl == "mongodb://root:example@mongo:27017/":
				{
					EPUtils.EPLogger(fmt.Sprintf("-of option selected as monngo. Proceeding with default mongo URL: %v. If you intend to connect to another url, please specify it with -murl option.", *ESPluginMongoUrl))
				}
			}
			if needsExit {
				printBasicInstruction(ESPluginFs, kibanaPluginFs, reportGeneratePluginFs, reportViewPluginFs)
				os.Exit(1)
			}
		}
	case kibanaPluginFs.Parsed():
		{
			needsExit := false

			needsExit = EPUtils.ValidateStringFlag(
				*kibanaPluginInputFilePath,
				"",
				"-f",
			) ||
				EPUtils.ValidatePositiveInt(
					*kibanaPluginThreadsNum,
					"-t",
				) ||
				EPUtils.ValidatePositiveInt(
					*kibanaPluginMaxIndexSizeToCollect,
					"-max-is",
				) ||
				EPUtils.ValidatePositiveInt(
					*kibanaPluginMaxIndicesToCollect,
					"-max-i",
				)

			if needsExit {
				printBasicInstruction(ESPluginFs, kibanaPluginFs, reportGeneratePluginFs, reportViewPluginFs)
				os.Exit(1)
			}
		}
	case reportGeneratePluginFs.Parsed():
		{
			needsExit := false

			needsExit = EPUtils.ValidateStringFlag(*reportGeneratePluginCollectionName, "", "-cn")

			if *reportGeneratePluginCollectionName != "elasticsearch" && *reportGeneratePluginCollectionName != "kibana" {
				EPUtils.EPLogger(fmt.Sprintf("-cn option: \"%v\" is not a valid collection name. should be either \"kibana\" or \"elasticsearch\".", *reportGeneratePluginCollectionName))
				needsExit = true
			}

			if needsExit {
				printBasicInstruction(ESPluginFs, kibanaPluginFs, reportGeneratePluginFs, reportViewPluginFs)
				os.Exit(1)
			}
		}
	case reportViewPluginFs.Parsed():
		{
			needsExit := false

			if needsExit {
				printBasicInstruction(ESPluginFs, kibanaPluginFs, reportGeneratePluginFs, reportViewPluginFs)
				os.Exit(1)
			}
		}
	default:
		{
			// should never get here
			panic("not a valid plugin name")
		}
	}

	return &Elasticpwn{
		mode: EPMode,
		elasticSearchPlugin: &EPPlugins.ElasticSearchPlugin{
			InputFilePath:        *ESPluginInputFilePath,
			ThreadsNum:           *ESPluginThreadsNum,
			OutputMode:           *ESPluginOutputMode,
			OutputFilePath:       *ESPluginOutputFilePath,
			MongoUrl:             *ESPluginMongoUrl,
			EsPluginMaxIndices:   *ESPluginMaxIndicesToCollect,
			EsPluginMaxIndexSize: *ESPluginMaxIndexSizeToCollect,
		},
		kibanaPlugin: &EPPlugins.KibanaPlugin{
			InputFilePath:  *kibanaPluginInputFilePath,
			ThreadsNum:     *kibanaPluginThreadsNum,
			OutputMode:     *kibanaPluginOutputMode,
			OutputFilePath: *kibanaPluginOutputFilePath,
			MongoUrl:       *kibanaPluginMongoUrl,
			MaxIndices:     *kibanaPluginMaxIndicesToCollect,
			MaxIndexSize:   *kibanaPluginMaxIndexSizeToCollect,
		},
		reportGeneratePlugin: &EPPlugins.ReportGeneratePlugin{
			MongoUrl:       *reportGeneratePluginMongoUrl,
			CollectionName: *reportGeneratePluginCollectionName,
			ServerRootUrl:  *reportGeneratePluginServerRootUrl,
		},
		reportViewPlugin: &EPPlugins.ReportViewPlugin{
			ReportDirectory: *reportViewPluginReportDirectory,
			ServeAtPort:     *reportViewPluginServeAtPort,
		},
	}
}

func setupGracefulExit(ep *Elasticpwn) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		EPUtils.EPLogger("Interrupted with Ctrl+C. Finalizing..")
		go func() {
			for {
				time.Sleep(time.Duration(2 * time.Second))
				EPUtils.EPLogger("Interrupted with Ctrl+C. Finalizing..")
			}
		}()
		switch {
		case ep.mode == "kibana" && ep.kibanaPlugin.OutputMode == "json":
			EPUtils.ConvertJSONObjectsToJSONArray(ep.kibanaPlugin.OutputFilePath)
		case ep.mode == "elasticsearch" && ep.kibanaPlugin.OutputMode == "json":
			EPUtils.ConvertJSONObjectsToJSONArray(ep.elasticSearchPlugin.OutputFilePath)
		}
		os.Exit(0)
	}()
}

func main() {
	defer os.Exit(0)

	elasticpwn := initializeElasticpwn()
	setupGracefulExit(elasticpwn)

	switch elasticpwn.mode {
	case "elasticsearch":
		{
			urls, elasticSearchCollection := EPPlugins.Prepare(
				elasticpwn.elasticSearchPlugin.InputFilePath,
				elasticpwn.elasticSearchPlugin.MongoUrl,
				elasticpwn.elasticSearchPlugin.OutputMode,
				"elasticsearch",
			)
			elasticpwn.elasticSearchPlugin.Run(urls, elasticSearchCollection)
			elasticpwn.elasticSearchPlugin.PostProcess()
			break
		}
	case "kibana":
		{
			urls, kibanaCollection := EPPlugins.Prepare(
				elasticpwn.kibanaPlugin.InputFilePath,
				elasticpwn.kibanaPlugin.MongoUrl,
				elasticpwn.kibanaPlugin.OutputMode,
				"kibana",
			)
			elasticpwn.kibanaPlugin.Run(urls, kibanaCollection)
			elasticpwn.kibanaPlugin.PostProcess()
			break
		}
	case "report":
		{
			if len(os.Args) <= 2 {
				log.Fatal("report mode is not defined. it should be either \"generate\" or \"view\".")
				os.Exit(1)
			}
			EPReportMode := os.Args[2]
			switch EPReportMode {
			case "generate":
				elasticpwn.reportGeneratePlugin.Run()
			case "view":
				elasticpwn.reportViewPlugin.Run()
			}
		}
	}
}
