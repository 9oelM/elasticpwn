package EPPlugins

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	EPUtils "github.com/9oelM/elasticpwn/elasticpwn/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// elastic-util contains common stuffs relevant to elastic products,
// i.e. elasticsearch and kibana.

// {
//     "health" : "yellow",
//     "status" : "open",
//     "index" : ".kibana",
//     "uuid" : "J399dVXVTDKTOShf6Q1DyA",
//     "pri" : "1",
//     "rep" : "1",
//     "docs.count" : "3",
//     "docs.deleted" : "0",
//     "store.size" : "54.1kb",
//     "pri.store.size" : "54.1kb"
// },
type IndexInfo struct {
	InterestingIndexInfo
	Health string `bson:"health,omitempty" json:"health"`
	Status string `bson:"status,omitempty" json:"status"`
	Uuid   string `bson:"uuid,omitempty" json:"uuid"`
	Pri    string `bson:"pri,omitempty" json:"pri"`
	Rep    string `bson:"rep,omitempty" json:"rep"`
}

type InterestingIndexInfo struct {
	Index        string `bson:"index,omitempty" json:"index"`
	DocsCount    string `bson:"docs.count,omitempty" json:"docs.count"`
	DocsDeleted  string `bson:"docs.deleted,omitempty" json:"docs.deleted"`
	StoreSize    string `bson:"store.size,omitempty" json:"store.size"`
	PriStoreSize string `bson:"pri.store.size,omitempty" json:"pri.store.size"`
}

func filterInterestingIndexFields(indexInfo IndexInfo) InterestingIndexInfo {
	var interestingIndexInfo InterestingIndexInfo

	interestingIndexInfo.DocsCount = indexInfo.DocsCount
	interestingIndexInfo.DocsDeleted = indexInfo.DocsDeleted
	interestingIndexInfo.StoreSize = indexInfo.StoreSize
	interestingIndexInfo.PriStoreSize = indexInfo.PriStoreSize
	interestingIndexInfo.Index = indexInfo.Index

	return interestingIndexInfo
}

func ProcessInterestingIndices(indices []IndexInfo) []InterestingIndexInfo {
	var interestingIndices []InterestingIndexInfo

	for _, indexInfo := range indices {
		if !IsUninterestingIndex(indexInfo.Index) {
			interestingIndices = append(interestingIndices, filterInterestingIndexFields(indexInfo))
		}
	}

	return interestingIndices
}

func NewInterstingInfoFromIndexSearch() *InterestingInfoFromIndexSearch {
	return &InterestingInfoFromIndexSearch{
		Emails:                []string{},
		Urls:                  []string{},
		PublicIPs:             []string{},
		MoreThanTwoDotsInName: []string{},
	}
}

func appendObjectsOfInterest(
	interestingInfoFromIndexSearch *InterestingInfoFromIndexSearch,
	interestingWordsFromIndexSearch []string,
	interestingInfo *InterestingInfoFromIndexSearch,
	interestingWords *[]string,
) {
	if interestingInfoFromIndexSearch != nil {
		interestingInfo.Emails = append(interestingInfo.Emails, interestingInfoFromIndexSearch.Emails...)
		interestingInfo.Urls = append(interestingInfo.Urls, interestingInfoFromIndexSearch.Urls...)
		interestingInfo.PublicIPs = append(interestingInfo.PublicIPs, interestingInfoFromIndexSearch.PublicIPs...)
		interestingInfo.MoreThanTwoDotsInName = append(interestingInfo.MoreThanTwoDotsInName, interestingInfoFromIndexSearch.MoreThanTwoDotsInName...)
	}
	if interestingWordsFromIndexSearch != nil {
		// modify the reference
		*interestingWords = append(*interestingWords, interestingWordsFromIndexSearch...)
	}
}

func ProcessInterestingInfoAndWordsThreadSafely(
	mu *sync.Mutex,
	singleInstanceScanResult interface{},
	indexInfoObjectInJsonString string,
) {
	interestingInfoFromIndexSearch := ProcessInterestingInfoFromIndexSearch(indexInfoObjectInJsonString)
	interestingWordsFromIndexSearch := ProcessInterestingWordsFromIndexSearch(indexInfoObjectInJsonString)

	// just bear with me, gopls does not officially support generics yet
	mu.Lock()
	defer mu.Unlock()
	switch scanResult := singleInstanceScanResult.(type) {
	case *SingleElasticsearchInstanceScanResult:
		if scanResult.InterestingInfo == nil {
			scanResult.InterestingInfo = NewInterstingInfoFromIndexSearch()
		}
		appendObjectsOfInterest(
			interestingInfoFromIndexSearch,
			interestingWordsFromIndexSearch,
			scanResult.InterestingInfo,
			&scanResult.InterestingWords,
		)
	case *SingleKibanaInstanceScanResult:
		if scanResult.InterestingInfo == nil {
			scanResult.InterestingInfo = NewInterstingInfoFromIndexSearch()
		}
		appendObjectsOfInterest(
			interestingInfoFromIndexSearch,
			interestingWordsFromIndexSearch,
			scanResult.InterestingInfo,
			&scanResult.InterestingWords,
		)
	default:
		panic(fmt.Sprintf("unrecognized type of scan result detected: %v\n", scanResult))
	}
}

func CheckOverGBIndexExistence(interestingIndices []InterestingIndexInfo) bool {
	hasOverGBIndex := false

	for _, indexInfo := range interestingIndices {
		hasOverGBIndex = strings.HasSuffix(indexInfo.StoreSize, "gb") ||
			strings.HasSuffix(indexInfo.StoreSize, "tb") ||
			strings.HasSuffix(indexInfo.StoreSize, "pb") ||
			strings.HasSuffix(indexInfo.PriStoreSize, "gb") ||
			strings.HasSuffix(indexInfo.PriStoreSize, "tb") ||
			strings.HasSuffix(indexInfo.PriStoreSize, "pb")

		if hasOverGBIndex {
			break
		}
	}

	return hasOverGBIndex
}

// @todo read from a dictionary file instead
var InterestingWordsLowercase = []string{
	// "appid",
	// "appname",
	"app",
	"domain",
	"referrer",
	"referer",
	"url",
	"phone",
	"address",
	"email",
	"transfer",
	"balance",
	"payment",
	"user",
	"username",
	"password",
	"token",
	"chat",
	"message",
}

var UninterstingWordsLowercase = []string{
	"example.com",
	"abc.com",
}

// if an index name is any one of these, it will very likely contain no useful information
var UninterstingIndexNamesLowercase = []string{
	// already hacked by 'meow' bot. see https://www.elastic.co/blog/protect-your-elasticsearch-deployments-against-meow-bot-attacks-for-free
	"meow",
	// elasticsearch internal index. useless.
	".geoip_databases",
	// hackers create an index like this when they find an open elasticsearch instance. therefore useless.
	"readme",
	"read_me",
	"read__me",
	// magento store deployment. usually nothing interesting
	"magento",
	"market.kline",
	"waveland-datas",
	"resources_index",
	"company-datas",
	"movies",
	// these indices come from frameworks and usually contain very useless info
	"zend3",
	"casa",
	"kkrp",
	"actuator",
	"m.api",
	"solr",
	"minio",
	"daman",
	"index.php",
	"index.js",
	"index.jsp",
	"index.py",
	"index.do",
	"index.htm",
	"index.html",
	"index.cfm",
	"index.aspx",
	"index.cgi",
	"index.pl",
	"index.asp",
	"index.action",
	"yz.jsp",
	"ilm-history",
	"seismic",
	"result-logs",
	// sometimes users just wanna test it out. useless
	"demo",
	// elasticsearch internal
	".async-search",
	// shopify plugins
	"vue_storefront",
	"produtos",
}

// must be coming from a framework. ditch all data if detected
var UninterstingIfAllOfTheseInIndices = []string{
	"service",
	"actuator",
	"casa",
	"auth",
}

var UninterestingIfExactlyMatches = []string{
	"service",
	"auth",
	"actions",
	"casa",
	"website",
	"api",
	"login",
	"config",
	"oauth",
	"connect",
	"v1",
	"v2",
	"km.asmx",
	"biz",
	// edx open source
	"wap",
	// edx open source
	"courseware_index",
	// some weird chinese data
	"v3",
	// some weird chinese data
	"video_info",
	// elasticsearch internal
	".elastichq",
	// trading related
	"btc.bitfinex.ticker",
	"eth.bitfinex.ticker",
	"btc.bitmex.ticker",
}

// if an index name starts with this, the index info is likely to be quite useless
var UninterstingIndexNamesStartingWith = []string{
	// elasticsearch internal
	".kibana",
	// elasticserach internal
	".apm",
	// elasticsearch index lifecycle
	"ilm-history",
}

// returns true if an index name is uninteresting, therefore useless to be inspected or stored
func IsUninterestingIndex(indexName string) bool {
	return EPUtils.ContainsStartsWith(indexName, UninterstingIndexNamesStartingWith) != -1 ||
		EPUtils.Contains(indexName, UninterstingIndexNamesLowercase) != -1 ||
		EPUtils.ContainsExactlyMatchesWith(indexName, UninterestingIfExactlyMatches) != -1
}

type InterestingInfoFromIndexSearch struct {
	Emails                []string `bson:"emails,omitempty" json:"emails"`
	Urls                  []string `bson:"urls,omitempty" json:"urls"`
	PublicIPs             []string `bson:"publicIps,omitempty" json:"publicIps"`
	MoreThanTwoDotsInName []string `bson:"moreThanTwoDotsInName,omitempty" json:"moreThanTwoDotsInName"`
}

func ProcessInterestingInfoFromIndexSearch(rawIndexSearchStringResult string) *InterestingInfoFromIndexSearch {

	return &InterestingInfoFromIndexSearch{
		Urls:                  EPUtils.FindAllUniqueUrls(rawIndexSearchStringResult),
		PublicIPs:             EPUtils.FindAllUniquePublicIps(rawIndexSearchStringResult),
		Emails:                EPUtils.Unique(EPUtils.EmailRegex.FindAllString(rawIndexSearchStringResult, -1)),
		MoreThanTwoDotsInName: EPUtils.Unique((EPUtils.AtLeastTwoDotsInSentenceRegex.FindAllString(rawIndexSearchStringResult, -1))),
	}
}

type InterestingWordRegex struct {
	regex *regexp.Regexp
	word  string
}

type InterestingWordRegexMatches struct {
	Matches []string `bson:"matches,omitempty" json:"matches"`
	Word    string   `bson:"word,omitempty" json:"word"`
}

var InterestingWordsRegexes = func() []*InterestingWordRegex {
	var regexes []*InterestingWordRegex

	for _, interestingWordLowerCase := range InterestingWordsLowercase {
		// https://stackoverflow.com/questions/15326421/how-do-i-do-a-case-insensitive-regular-expression-in-go
		// intended to match things like "transfer: 100 USD"
		// (?i) = ignore case
		// ^ = start of the word
		// (.) = anything except line break
		// {1,30} = between 1 and 10 times
		regex := regexp.MustCompile(fmt.Sprintf("(?i)(%s)(.){1,30}", interestingWordLowerCase))
		regexes = append(regexes, &InterestingWordRegex{
			regex: regex,
			word:  interestingWordLowerCase,
		})
	}

	return regexes
}()

func ProcessInterestingWordsFromIndexSearch(rawIndexSearchStringResult string) []string {
	var matches []string
	for _, interestingWordRegex := range InterestingWordsRegexes {
		matchesForSingleRegex := EPUtils.Unique(interestingWordRegex.regex.FindAllString(rawIndexSearchStringResult, -1))
		matches = append(matches, matchesForSingleRegex...)
	}

	return matches
}

func AppendScanResultToJSONFileWithNewline(outputFilePath string, marshalledScanResult []byte, mu *sync.Mutex) {
	// I/O should be thread-safe already, but just be extra safe
	mu.Lock()
	defer mu.Unlock()
	f, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		EPUtils.EPLogger(fmt.Sprintf("Failed to create/open %v for output.\n", outputFilePath))
	}

	singleKibanaInstanceScanResultMarshalledWithCommaAndNewline := append(marshalledScanResult, ",\n"...)
	if _, err := f.Write(singleKibanaInstanceScanResultMarshalledWithCommaAndNewline); err != nil {
		EPUtils.EPLogger(fmt.Sprintf("Failed to write to %v.\n", outputFilePath))
	}
	if err := f.Close(); err != nil {
		EPUtils.EPLogger(fmt.Sprintf("Failed to close i/o from %v.\n", outputFilePath))
	}
}

func Prepare(
	inputFilePath string,
	mongoUrl string,
	outputMode string,
	mongoCollectionName string,
) (urls []string, elasticSearchCollection *mongo.Collection) {
	urls = strings.Split(EPUtils.ReadUrlsFromFile(inputFilePath), "\n")
	EPUtils.EPLogger(fmt.Sprintf("Total %d lines of URLs detected\n", len(urls)))
	if outputMode == "mongo" {
		mongoClient := EPUtils.InitMongoConnnection(mongoUrl)
		// @todo customize db/collection name
		elasticSearchCollection = mongoClient.Database("ep").Collection(mongoCollectionName)
	}

	return urls, elasticSearchCollection
}

func InsertSingleScanResultToMongo(
	collection *mongo.Collection,
	singleScanResult interface{},
	rootUrl string,
) {
	insertContext, cancelInsert := context.WithTimeout(context.Background(),
		10*time.Second)
	insertResult, insertResultErr := collection.InsertOne(insertContext, bson.M{"scanResult": singleScanResult})

	defer cancelInsert()
	if insertResultErr != nil {
		EPUtils.EPLogger(fmt.Sprintf("Error while inserting info about %s into MongoDB.", rootUrl))
		fmt.Println(insertResultErr)
	} else {
		EPUtils.EPLogger(fmt.Sprintf("Inserted info about %s into MongoDB Collection. ID: %s", rootUrl, insertResult.InsertedID))
	}
}
