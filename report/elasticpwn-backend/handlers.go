package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func isValidUrl(rawUrl string) bool {
	_, err := url.ParseRequestURI(rawUrl)

	foundIp := IpWithOptionalPortRegex.FindAllString(rawUrl, -1)
	foundIpv6 := Ipv6WithOptionalPortRegex.FindAllString(rawUrl, -1)
	return err == nil || foundIp != nil || foundIpv6 != nil
}

var DefaultOrdered = false

func PingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func ErrorHandler(c *gin.Context, err error) {
	c.JSON(500, gin.H{
		"error": fmt.Sprintf("%v", err),
	})
}

var IpWithOptionalPortRegex = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}(:[0-9]+)?`)

var Ipv6WithOptionalPortRegex = regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))(:[0-9]+)?`)

type PostUrlsHandlerPayload struct {
	Urls []string `form:"urls" json:"urls" xml:"urls"  binding:"required"`
}

func PostUrlsHandlerGenerator(urlsCollection *mongo.Collection) func(c *gin.Context) {
	return func(c *gin.Context) {
		var postUrlsHandlerPayload PostUrlsHandlerPayload

		if err := c.ShouldBindJSON(&postUrlsHandlerPayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "should include urls"})
		}

		log.Println(fmt.Sprintf("received %v", postUrlsHandlerPayload.Urls))

		var urlsBsonDArray []interface{}
		for _, url := range postUrlsHandlerPayload.Urls {
			if !isValidUrl(url) {
				c.JSON(http.StatusNotAcceptable, gin.H{
					"error": fmt.Sprintf("%v is not a valid url", url),
				})
				return
			} else {
				urlsBsonDArray = append(urlsBsonDArray, bson.M{"url": url})
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := urlsCollection.InsertMany(
			ctx,
			urlsBsonDArray,
			// just make it idempotent by ignoring duplicate key error
			&options.InsertManyOptions{Ordered: &DefaultOrdered},
		)

		// just make it idempotent by ignoring duplicate key error
		if err != nil && !mongo.IsDuplicateKeyError(err) {
			ErrorHandler(c, err)
			return
		}

		c.JSON(200, gin.H{
			"error": nil,
		})
	}
}

func GetUrlsHandlerGenerator(urlsCollection *mongo.Collection) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		distinct, err := urlsCollection.Distinct(ctx, "url", bson.M{})
		if err != nil {
			ErrorHandler(c, err)
			return
		}

		c.JSON(200, gin.H{
			"urls": distinct,
		})
	}
}

type DeleteUrlsHandlerPayload struct {
	Collection string `form:"collection" json:"collection" xml:"collection"  binding:"required"`
}

// only for the case where you delete all urls in a collection (for cleaning up test data, etc)
func DeleteUrlsHandlerGenerator(mongoClient *mongo.Client, databaseName string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var deleteUrlsHandlerPayload DeleteUrlsHandlerPayload

		if err := c.ShouldBindJSON(&deleteUrlsHandlerPayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "should include collection"})
			return
		}

		ctx0, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		db := mongoClient.Database(databaseName)
		collectionNames, err := db.ListCollectionNames(ctx0, bson.D{})
		if err != nil {
			ErrorHandler(c, err)
			return
		}

		existsCollectionName := false
		for _, collectionName := range collectionNames {
			if collectionName == deleteUrlsHandlerPayload.Collection {
				existsCollectionName = true
				break
			}
		}

		if !existsCollectionName {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("collection name should be any of: %v", collectionNames),
			})
			return
		}

		ctx1, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = db.Collection(deleteUrlsHandlerPayload.Collection).Drop(ctx1)

		if err != nil {
			ErrorHandler(c, err)
			return
		}

		c.JSON(200, gin.H{
			"error": nil,
		})
	}
}
