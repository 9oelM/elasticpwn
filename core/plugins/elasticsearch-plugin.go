package EPPlugins

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	EPUtils "github.com/9oelm/elasticpwn/core/util"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ElasticSearchPlugin struct {
	InputFilePath        string
	ThreadsNum           int
	OutputMode           string
	OutputFilePath       string
	MongoUrl             string
	EsPluginMaxIndices   int
	EsPluginMaxIndexSize int
}

// all jsons but in a stringified form
type SingleElasticsearchInstanceScanResult struct {
	Id          primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	RootUrl     string                 `bson:"rootUrl,omitempty" json:"rootUrl"`
	Indices     []InterestingIndexInfo `bson:"indices,omitempty" json:"indices"`
	IndicesInfo sync.Map               `bson:"indicesInfo,omitempty" json:"indicesInfo"`
	// sync.Map can't be (un)marshalled
	IndicesInfoInJson            map[string]interface{}          `bson:"indicesInfoInJson,omitempty" json:"indicesInfoInJson"`
	InterestingWords             []string                        `bson:"interestingWords,omitempty" json:"interestingWords"`
	InterestingInfo              *InterestingInfoFromIndexSearch `bson:"interestingInfo,omitempty" json:"interestingInfo"`
	HasAtLeastOneIndexSizeOverGB bool                            `bson:"hasAtLeastOneIndexSizeOverGB,omitempty" json:"hasAtLeastOneIndexSizeOverGB"`
	Aliases                      []interface{}                   `bson:"aliases,omitempty" json:"aliases"`
	Allocations                  []interface{}                   `bson:"allocations,omitempty" json:"allocations"`
	Nodes                        []interface{}                   `bson:"nodes,omitempty" json:"nodes"`
	CreatedAt                    time.Time                       `bson:"created_at" json:"created_at"`
	IsInitialized                bool                            `bson:"isInitialized,omitempty" json:"isInitialized"`
}

const (
	API_INDICES     = "/_cat/indices"
	API_ALIASES     = "/_cat/aliases"
	API_ALLOCATIONS = "/_cat/allocation"
	API_NODES       = "/_cat/nodes"
	API_SEARCH      = "/_search"

	// unused for now because they usually contain quite useless info from a security perspective
	API_COUNT          = "/_cat/count"
	API_MASTER         = "/_cat/master"
	API_NODE_ATTRS     = "/_cat/nodeattrs"
	API_PENDING_TASKS  = "/_cat/pending_tasks"
	API_PLUGINS        = "/_cat/plugins"
	API_TASKS          = "/_cat/tasks"
	API_TRAINED_MODELS = "/_cat/ml/trained_models"
	API_TRANSFORMS     = "/_cat/transforms"
)

const (
	Q_FORMAT_JSON = "format=json"
	Q_SIZE_X      = "size={INDEX_SIZE}"
)

var Q_SIZE_X_FORMAT_JSON = fmt.Sprintf(strings.Join([]string{
	Q_FORMAT_JSON,
	Q_SIZE_X,
}, "&"))

// any better way?
func elasticSearchResultSwitch(singleElasticsearchInstanceScanResult *SingleElasticsearchInstanceScanResult, endpoint string, resp string) {
	if strings.Contains(endpoint, API_INDICES) {
		var jsonArrayResponse []IndexInfo
		jsonUnmarshalErr := json.Unmarshal([]byte(resp), &jsonArrayResponse)

		if jsonUnmarshalErr != nil {
			EPUtils.EPLogger(fmt.Sprintf("Error while unmarshalling %s", resp))
			return
		}
		singleElasticsearchInstanceScanResult.Indices = ProcessInterestingIndices(jsonArrayResponse)

		return
	}

	var jsonArrayResponse []interface{}
	jsonUnmarshalErr := json.Unmarshal([]byte(resp), &jsonArrayResponse)

	if jsonUnmarshalErr != nil {
		EPUtils.EPLogger(fmt.Sprintf("Error while unmarshalling %s", resp))
	}

	if jsonUnmarshalErr != nil {
		return
	}
	// []interface{}

	switch {
	case strings.Contains(endpoint, API_ALIASES):
		{
			singleElasticsearchInstanceScanResult.Aliases = jsonArrayResponse
		}
	case strings.Contains(endpoint, API_ALLOCATIONS):
		{
			singleElasticsearchInstanceScanResult.Allocations = jsonArrayResponse
		}
	case strings.Contains(endpoint, API_NODES):
		{
			singleElasticsearchInstanceScanResult.Nodes = jsonArrayResponse
		}
	}
}

func (elasticSearchPlugin *ElasticSearchPlugin) requestAllAPIs(url string, singleElasticsearchInstanceScanResult *SingleElasticsearchInstanceScanResult) {
	endpoints := []string{
		// just make sure you requeset all indices
		fmt.Sprintf("%s?%s&size=1000", API_INDICES, Q_FORMAT_JSON),
		fmt.Sprintf("%s?%s", API_ALIASES, Q_FORMAT_JSON),
		fmt.Sprintf("%s?%s", API_ALLOCATIONS, Q_FORMAT_JSON),
		fmt.Sprintf("%s?%s", API_NODES, Q_FORMAT_JSON),
	}
	wg := sync.WaitGroup{}
	var validHTTPRequestCount EPUtils.Count32 = 0
	for _, endpoint := range endpoints {
		wg.Add(1)
		go func(endpoint string) {
			defer wg.Done()
			var (
				resp string
				err  error
			)
			var finalUrl = fmt.Sprintf("%s%s", url, endpoint)
			EPUtils.EPLogger(fmt.Sprintf("Requesting %s\n", finalUrl))
			resp, _, err = EPUtils.SendFailSafeHTTPRequest(finalUrl, 15, false, map[string]string{}, "GET")
			if err != nil {
				return
			}
			elasticSearchResultSwitch(singleElasticsearchInstanceScanResult, endpoint, resp)
			validHTTPRequestCount.Inc()
		}(endpoint)

	}
	wg.Wait()
	singleElasticsearchInstanceScanResult.IsInitialized = validHTTPRequestCount > 0
}

// outputs something like 123.123.123.123:9200/index_name/_search?format=json&size=1000
func (elasticSearchPlugin *ElasticSearchPlugin) buildElasticSearchIndexSearchAPI(rootUrl string, indexName string) string {
	queryString := strings.ReplaceAll(Q_SIZE_X_FORMAT_JSON, `{INDEX_SIZE}`, fmt.Sprintf("%v", elasticSearchPlugin.EsPluginMaxIndexSize))
	return fmt.Sprintf(
		"%s/%s%s?%s",
		rootUrl,
		indexName,
		API_SEARCH,
		queryString,
	)
}

func (elasticSearchPlugin *ElasticSearchPlugin) searchSingleIndexInfo(
	indexName string,
	singleElasticsearchInstanceScanResult *SingleElasticsearchInstanceScanResult,
) string {
	getIndexEndpoint := elasticSearchPlugin.buildElasticSearchIndexSearchAPI(singleElasticsearchInstanceScanResult.RootUrl, indexName)
	EPUtils.EPLogger(fmt.Sprintf("Requesting %s", getIndexEndpoint))
	searchIndexResult, _, searchIndexResultErr := EPUtils.SendFailSafeHTTPRequest(getIndexEndpoint, 30, false, map[string]string{}, "GET")
	if searchIndexResultErr != nil {
		EPUtils.EPLogger(fmt.Sprintf("Error in requesting index %s from %s", indexName, singleElasticsearchInstanceScanResult.RootUrl))

		return ""
	}
	var jsonArrayResponse interface{}
	jsonUnmarshalErr := json.Unmarshal([]byte(searchIndexResult), &jsonArrayResponse)
	if jsonUnmarshalErr != nil {
		EPUtils.EPLogger(fmt.Sprintf("Error while unmarshalling result from %s", getIndexEndpoint))
		fmt.Println(searchIndexResultErr)
		return ""
	}

	// each index is unique
	// https://stackoverflow.com/questions/45585589/golang-fatal-error-concurrent-map-read-and-map-write/45585833
	singleElasticsearchInstanceScanResult.IndicesInfo.Store(indexName, jsonArrayResponse)
	return searchIndexResult
}

func (elasticSearchPlugin *ElasticSearchPlugin) scanInterestingIndices(
	singleElasticsearchInstanceScanResult *SingleElasticsearchInstanceScanResult,
) {
	EPUtils.EPLogger(fmt.Sprintf("%s has valid indices. Will query them all", singleElasticsearchInstanceScanResult.RootUrl))

	mu := &sync.Mutex{}
	getIndicesWg := sync.WaitGroup{}
	concurrentGoroutines := make(chan struct{}, 5)
	var maxIndicesFromSingleElasticsearchInstanceScanResult []InterestingIndexInfo
	maxIndicesNum := elasticSearchPlugin.EsPluginMaxIndices
	if len(singleElasticsearchInstanceScanResult.Indices) > maxIndicesNum {
		EPUtils.EPLogger(fmt.Sprintf("%s has number of indices more than %d. Will only request first %d indices as specified in -max-i option", singleElasticsearchInstanceScanResult.RootUrl, maxIndicesNum, maxIndicesNum))
		maxIndicesFromSingleElasticsearchInstanceScanResult = singleElasticsearchInstanceScanResult.Indices[:maxIndicesNum]
	} else {
		maxIndicesFromSingleElasticsearchInstanceScanResult = singleElasticsearchInstanceScanResult.Indices
	}

	for _, indexInfo := range maxIndicesFromSingleElasticsearchInstanceScanResult {
		getIndicesWg.Add(1)
		go func(indexInfo InterestingIndexInfo) {
			defer getIndicesWg.Done()
			concurrentGoroutines <- struct{}{}
			// 123.123.123.123/example-index/_search?format=json&size=1000&pretty=true
			// don't store uninteresting index names
			searchIndexResult := elasticSearchPlugin.searchSingleIndexInfo(
				indexInfo.Index,
				singleElasticsearchInstanceScanResult,
			)

			if searchIndexResult != "" {
				ProcessInterestingInfoAndWordsThreadSafely(mu, singleElasticsearchInstanceScanResult, searchIndexResult)
			}
			<-concurrentGoroutines
		}(indexInfo)
	}
	getIndicesWg.Wait()
	EPUtils.EPLogger(fmt.Sprintf("Finished querying indices information for %s", singleElasticsearchInstanceScanResult.RootUrl))
}

func (elasticSearchPlugin *ElasticSearchPlugin) scanSingleElasticsearchInstance(url string) *SingleElasticsearchInstanceScanResult {
	_, _, errFromRootUrl := EPUtils.SendFailSafeHTTPRequest(url, 10, true, map[string]string{}, "GET")

	if errFromRootUrl != nil {
		return &SingleElasticsearchInstanceScanResult{
			Id:            primitive.NewObjectID(),
			CreatedAt:     time.Now(),
			IsInitialized: false,
			RootUrl:       url,
		}
	} else {
		EPUtils.EPLogger(fmt.Sprintf("%s is a working elasticSearch instance\n", url))
	}
	singleElasticsearchInstanceScanResult := &SingleElasticsearchInstanceScanResult{
		Id:          primitive.NewObjectID(),
		CreatedAt:   time.Now(),
		IndicesInfo: sync.Map{},
		RootUrl:     url,
	}
	elasticSearchPlugin.requestAllAPIs(url, singleElasticsearchInstanceScanResult)

	if singleElasticsearchInstanceScanResult.Indices == nil {
		EPUtils.EPLogger(fmt.Sprintf("Failed to get indices from %v\n", url))
		return singleElasticsearchInstanceScanResult
	}
	singleElasticsearchInstanceScanResult.HasAtLeastOneIndexSizeOverGB = CheckOverGBIndexExistence(singleElasticsearchInstanceScanResult.Indices)
	elasticSearchPlugin.scanInterestingIndices(singleElasticsearchInstanceScanResult)

	return singleElasticsearchInstanceScanResult
}

var elasticSearchPluginFileWriteMutex sync.Mutex

func (elasticSearchPlugin *ElasticSearchPlugin) outputSingleElasticsearchInstanceScanResult(singleElasticsearchInstanceScanResult *SingleElasticsearchInstanceScanResult, elasticSearchCollection *mongo.Collection) {
	// sync.Map can't be (un)marshalled
	singleElasticsearchInstanceScanResult.IndicesInfoInJson = EPUtils.ConvertSyncMapToMap(singleElasticsearchInstanceScanResult.IndicesInfo)
	singleElasticsearchInstanceScanResultMarshalled, singleElasticsearchInstanceScanResultMarshalErr := json.Marshal(singleElasticsearchInstanceScanResult)

	if singleElasticsearchInstanceScanResultMarshalErr != nil {
		EPUtils.EPLogger(fmt.Sprintf("Error while marshalling info about %s\n", singleElasticsearchInstanceScanResult.RootUrl))
		fmt.Println(singleElasticsearchInstanceScanResultMarshalErr)
	}

	switch elasticSearchPlugin.OutputMode {
	case "json", "plain":
		{
			AppendScanResultToJSONFileWithNewline(elasticSearchPlugin.OutputFilePath, singleElasticsearchInstanceScanResultMarshalled, &elasticSearchPluginFileWriteMutex)
		}
	case "mongo":
		{
			if elasticSearchCollection == nil {
				panic("mongo option was specified but elasticSearchCollection is nil")
			}
			InsertSingleScanResultToMongo(elasticSearchCollection, singleElasticsearchInstanceScanResult, singleElasticsearchInstanceScanResult.RootUrl)

		}
	}
}

func (elasticSearchPlugin *ElasticSearchPlugin) Run(urls []string, elasticSearchCollection *mongo.Collection) {
	var finishedGoRoutineCount EPUtils.Count32
	concurrentGoroutines := make(chan struct{}, elasticSearchPlugin.ThreadsNum)
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string, elasticSearchCollection *mongo.Collection) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			singleElasticsearchInstanceScanResult := elasticSearchPlugin.scanSingleElasticsearchInstance(url)
			elasticSearchPlugin.outputSingleElasticsearchInstanceScanResult(singleElasticsearchInstanceScanResult, elasticSearchCollection)
			finishedGoRoutineCount.Inc()
			<-concurrentGoroutines
		}(url, elasticSearchCollection)
	}
	go func() {
		for {
			time.Sleep(time.Duration(2 * time.Second))
			EPUtils.EPLogger(fmt.Sprintf("%d/%d of URLs completed\n", finishedGoRoutineCount.Get(), len(urls)))
		}
	}()
	wg.Wait()
	EPUtils.EPLogger(fmt.Sprintf("%d/%d of URLs completed\n", len(urls), len(urls)))
	EPUtils.EPLogger("Scan finished")
}

func (elasticSearchPlugin *ElasticSearchPlugin) PostProcess() {
	if elasticSearchPlugin.OutputMode != "json" {
		return
	}
	EPUtils.ConvertJSONObjectsToJSONArray(elasticSearchPlugin.OutputFilePath)
}
