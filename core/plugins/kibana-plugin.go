package EPPlugins

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	EPLookup_addrs "github.com/9oelM/elasticpwn/core/lookup-addrs"
	EPUtils "github.com/9oelM/elasticpwn/core/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Scraps data from an open Kibana instance using webdriver
type KibanaPlugin struct {
	InputFilePath  string
	ThreadsNum     int
	OutputMode     string
	OutputFilePath string
	MongoUrl       string
	MaxIndices     int
	MaxIndexSize   int
}

type IpInfo struct {
	SubjectUrls   string `bson:"subjectUrls,omitempty" json:"subjectUrls"`
	Organizations string `bson:"organizations,omitempty" json:"organizations"`
	// ex. "amazonaws.com."
	CloudHostingProvider string `bson:"cloudHostingProvider,omitempty" json:"cloudHostingProvider"`
	Cname                string `bson:"cname,omitempty" json:"cname"`
}

// The interesting data that was stored in a single Kibana instance
type SingleKibanaInstanceScanResult struct {
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	// indices string
	IsInitialized bool `bson:"isInitialized,omitempty" json:"isInitialized"`
	// example: 123.123.123.123:5000
	RootUrl                      string `bson:"rootUrl,omitempty" json:"rootUrl"`
	IpInfo                       *IpInfo
	HasAtLeastOneIndexSizeOverGB bool `bson:"hasAtLeastOneIndexSizeOverGB,omitempty" json:"hasAtLeastOneIndexSizeOverGB"`
	// only stores indices of interesting names
	Indices     []InterestingIndexInfo `bson:"indices,omitempty" json:"indices"`
	IndicesInfo sync.Map
	// sync.Map can't be (un)marshalled
	IndicesInfoInJson map[string]interface{}          `bson:"indicesInfoInJson,omitempty" json:"indicesInfoInJson"`
	InterestingWords  []string                        `bson:"interestingWords,omitempty" json:"interestingWords"`
	InterestingInfo   *InterestingInfoFromIndexSearch `bson:"interestingInfo,omitempty" json:"interestingInfo"`
}

type KibanaRequests struct {
	GET_root    string
	GET_indices string
}

const CONSOLE_PATH = "app/kibana#/dev_tools/console?_g=()"

type KibanaGetRequests struct {
	indices     string
	indexSearch string
}

// some versions have slightly different APIs
type KibanaAPI struct {
	get *KibanaGetRequests
}

// Working versions:
// 5.2.1
var kibanaVer5_2_1 = &KibanaAPI{
	get: &KibanaGetRequests{
		indices:     "api/console/proxy?uri=_cat%2Findices%3Fformat%3Djson",
		indexSearch: "api/console/proxy?uri={INDEX_NAME}%2F_search%3Fformat%3Djson%26size%3D{INDEX_SIZE}",
	},
}

// Working versions:
// 5.6.16
// 6.2.4
// 6.2.2
// 7.15.0
var kibanaVer7_15_0 = &KibanaAPI{
	get: &KibanaGetRequests{
		//  "api/console/proxy?path=%2F_cat%2Findices%3Fformat%3Djson&method=GET"
		indices:     "api/console/proxy?path=%2F_cat%2Findices%3Fformat%3Djson&method=GET",
		indexSearch: "api/console/proxy?path=%2F{INDEX_NAME}%2F_search%3Fformat%3Djson%26size%3D{INDEX_SIZE}&method=GET",
	},
}

func (kpAPI *KibanaAPI) buildKibanaIndexSearchAPI(rootUrl string, indexName string, indexSize int) string {
	builtAPI := strings.Replace(kpAPI.get.indexSearch, `{INDEX_NAME}`, indexName, 1)
	builtAPI = strings.Replace(builtAPI, `{INDEX_SIZE}`, fmt.Sprintf("%v", indexSize), 1)

	return fmt.Sprintf("%s/%s", rootUrl, builtAPI)
}

var kibanaHeader = map[string]string{
	// kibana requires this useless header to be set: https://discuss.elastic.co/t/where-can-i-get-the-correct-kbn-xsrf-value-for-my-plugin-http-requests/158725
	"kbn-xsrf":     "_",
	"Content-Type": "application/json",
	"User-Agent":   "curl/7.0.0",
	"Accept":       "*/*",
}

// to see if we need to abort early because the instance is not up at all
// returns true if unhealthy
func (kp *KibanaPlugin) checkIsInstanceDown(rootUrl string) bool {
	// you need to insert kibana headers even for the index page
	anything, statusCode, _ := EPUtils.SendFailSafeHTTPRequest(rootUrl, 15, false, kibanaHeader, "GET")

	return anything == "" && statusCode != 200
}

// returns nil if could not get indices
func (kp *KibanaPlugin) getIndices(rootUrl string) []IndexInfo {
	EPUtils.EPLogger(fmt.Sprintf("%v has a working Kibana frontend", rootUrl))

	var indicesArray []IndexInfo
	allPossibleGetIndicesRequests := []string{
		fmt.Sprintf("%s/%s", rootUrl, kibanaVer5_2_1.get.indices),
		fmt.Sprintf("%s/%s", rootUrl, kibanaVer7_15_0.get.indices),
	}
	for _, req := range allPossibleGetIndicesRequests {
		var indicesArrayInJsonString string = ""
		var statusCode int
		switch {
		case strings.HasSuffix(req, kibanaVer7_15_0.get.indices):
			{
				// recent versions of kibana has this weird system where you need to POST in order to GET through proxy
				indicesArrayInJsonString, statusCode, _ = EPUtils.SendFailSafeHTTPRequest(req, 15, false, kibanaHeader, "POST")
				break
			}
		case strings.HasSuffix(req, kibanaVer5_2_1.get.indices):
			{
				indicesArrayInJsonString, statusCode, _ = EPUtils.SendFailSafeHTTPRequest(req, 15, false, kibanaHeader, "GET")
				break
			}
		}

		jsonUnmarshalErr := json.Unmarshal([]byte(indicesArrayInJsonString), &indicesArray)

		if jsonUnmarshalErr == nil && statusCode != 404 {
			break
		} else {
			indicesArray = nil
		}
	}

	return indicesArray
}

func (kp *KibanaPlugin) scanInterestingIndices(
	singleKibanaInstanceScanResult *SingleKibanaInstanceScanResult,
) {
	wg := sync.WaitGroup{}
	// setting this number high will likely cause a panic (too many files open) and high memory usage
	// because there could be many indices
	concurrentGoroutines := make(chan struct{}, 5)
	mu := &sync.Mutex{}

	maxIndicesNum := kp.MaxIndices
	if len(singleKibanaInstanceScanResult.Indices) <= maxIndicesNum {
		maxIndicesNum = len(singleKibanaInstanceScanResult.Indices)
	}

	// some low versions of kibana do not support /mget method (request multiple indices with one call), so just request each index respectively
	for _, indexInfo := range singleKibanaInstanceScanResult.Indices[:maxIndicesNum] {
		wg.Add(1)

		go func(indexInfo InterestingIndexInfo) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}

			var indexInfoObject map[string]interface{}
			allPossibleGetIndexSearchRequests := []string{
				kibanaVer7_15_0.buildKibanaIndexSearchAPI(singleKibanaInstanceScanResult.RootUrl, indexInfo.Index, kp.MaxIndexSize),
				kibanaVer5_2_1.buildKibanaIndexSearchAPI(singleKibanaInstanceScanResult.RootUrl, indexInfo.Index, kp.MaxIndexSize),
			}
			for _, req := range allPossibleGetIndexSearchRequests {
				var indexInfoObjectInJsonString string
				var statusCode int
				switch req {
				case allPossibleGetIndexSearchRequests[0]:
					{
						// recent versions of kibana has this weird system where you need to POST in order to GET through proxy
						// keep timeout reasonably low, otherwise will cause memory usage spike in low-end machines
						indexInfoObjectInJsonString, statusCode, _ = EPUtils.SendFailSafeHTTPRequest(req, 30, false, kibanaHeader, "POST")

						break
					}
				case allPossibleGetIndexSearchRequests[1]:
					{
						indexInfoObjectInJsonString, statusCode, _ = EPUtils.SendFailSafeHTTPRequest(req, 30, false, kibanaHeader, "GET")

						break
					}
				}

				jsonUnmarshalErr := json.Unmarshal([]byte(indexInfoObjectInJsonString), &indexInfoObject)

				if jsonUnmarshalErr == nil && strings.TrimSpace(indexInfoObjectInJsonString) != "" && statusCode != 404 {
					singleKibanaInstanceScanResult.IndicesInfo.Store(indexInfo.Index, indexInfoObject)

					ProcessInterestingInfoAndWordsThreadSafely(mu, singleKibanaInstanceScanResult, indexInfoObjectInJsonString)
					break
				} else {
					indexInfoObject = nil
				}
			}

			<-concurrentGoroutines
		}(indexInfo)
	}

	wg.Wait()
}

// rootUrl example: 123.123.123.123:5601
func (kp *KibanaPlugin) scanSingleKibanaInstance(rootUrl string) *SingleKibanaInstanceScanResult {
	singleKibanaInstanceScanResult := &SingleKibanaInstanceScanResult{
		IsInitialized: false,
		RootUrl:       rootUrl,
		Id:            primitive.NewObjectID(),
		CreatedAt:     time.Now(),
		IndicesInfo:   sync.Map{},
	}

	isInstanceDown := kp.checkIsInstanceDown(rootUrl)
	if isInstanceDown {
		EPUtils.EPLogger(fmt.Sprintf("%v is down\n", rootUrl))

		return singleKibanaInstanceScanResult
	}

	allIndices := kp.getIndices(rootUrl)
	if allIndices == nil {
		EPUtils.EPLogger(fmt.Sprintf("Failed to get indices from %v\n", rootUrl))
		return singleKibanaInstanceScanResult
	}

	interestingIndices := ProcessInterestingIndices(allIndices)
	if interestingIndices == nil {
		EPUtils.EPLogger(fmt.Sprintf("Failed to get interesting indices from %v\n", rootUrl))
		return singleKibanaInstanceScanResult
	}
	singleKibanaInstanceScanResult.HasAtLeastOneIndexSizeOverGB = CheckOverGBIndexExistence(interestingIndices)
	singleKibanaInstanceScanResult.IsInitialized = true

	singleKibanaInstanceScanResult.Indices = interestingIndices

	kp.scanInterestingIndices(singleKibanaInstanceScanResult)

	return singleKibanaInstanceScanResult
}

var kibanaPluginFileWriteMutex sync.Mutex

func (kp *KibanaPlugin) outputSingleKibanaInstanceScanResult(
	singleKibanaInstanceScanResult *SingleKibanaInstanceScanResult,
	kibanaCollection *mongo.Collection,
) {
	indicesInfoMap := EPUtils.ConvertSyncMapToMap(singleKibanaInstanceScanResult.IndicesInfo)
	singleKibanaInstanceScanResult.IndicesInfoInJson = indicesInfoMap

	singleKibanaInstanceScanResultMarshalled, singleKibanaInstanceScanResultMarshalErr := json.Marshal(singleKibanaInstanceScanResult)

	if singleKibanaInstanceScanResultMarshalErr != nil {
		EPUtils.EPLogger(fmt.Sprintf("Failed to marshall scan result for %v", singleKibanaInstanceScanResult.RootUrl))

		return
	}

	switch kp.OutputMode {
	case "json", "plain":
		{
			AppendScanResultToJSONFileWithNewline(kp.OutputFilePath, singleKibanaInstanceScanResultMarshalled, &kibanaPluginFileWriteMutex)
		}
	case "mongo":
		{
			if kibanaCollection == nil {
				panic("mongo option was given, but collection is nil")
			}
			InsertSingleScanResultToMongo(kibanaCollection, singleKibanaInstanceScanResult, singleKibanaInstanceScanResult.RootUrl)
		}
	default:
		{
			panic(fmt.Sprintf("Unrecognized output mode: %v", kp.OutputMode))
		}
	}
}

func (kp *KibanaPlugin) scanKibanaInstanceAndIpInfo(url string) *SingleKibanaInstanceScanResult {
	singleKibanaInstanceScanResultChan := make(chan *SingleKibanaInstanceScanResult)
	ipInfoChan := make(chan *IpInfo)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(url string) {
		singleKibanaInstanceScanResult := kp.scanSingleKibanaInstance(url)
		singleKibanaInstanceScanResultChan <- singleKibanaInstanceScanResult
	}(url)
	wg.Add(1)
	go func(url string) {
		cloudHostingProvider, subjectUrls, organizations, cname := EPLookup_addrs.GetIpInfo(url)
		ipInfoChan <- &IpInfo{
			CloudHostingProvider: cloudHostingProvider,
			SubjectUrls:          subjectUrls,
			Organizations:        organizations,
			Cname:                cname,
		}
	}(url)
	singleKibanaInstanceScanResult, ipInfo := <-singleKibanaInstanceScanResultChan, <-ipInfoChan
	singleKibanaInstanceScanResult.IpInfo = ipInfo

	return singleKibanaInstanceScanResult
}

func (kp *KibanaPlugin) Run(urls []string, kibanaCollection *mongo.Collection) {
	wg := sync.WaitGroup{}
	concurrentGoroutines := make(chan struct{}, kp.ThreadsNum)
	var finishedGoRoutineCount EPUtils.Count32

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			singleKibanaInstanceScanResult := kp.scanKibanaInstanceAndIpInfo(url)

			kp.outputSingleKibanaInstanceScanResult(singleKibanaInstanceScanResult, kibanaCollection)
			finishedGoRoutineCount.Inc()
			<-concurrentGoroutines
		}(url)
	}
	go func() {
		for {
			time.Sleep(time.Duration(1 * time.Second))
			EPUtils.EPLogger(fmt.Sprintf("%d/%d of URLs completed\n", finishedGoRoutineCount.Get(), len(urls)))
		}
	}()
	wg.Wait()
}

func (kp *KibanaPlugin) PostProcess() {
	if kp.OutputMode != "json" {
		return
	}
	EPUtils.ConvertJSONObjectsToJSONArray(kp.OutputFilePath)
}
