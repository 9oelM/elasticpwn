package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func insertTestUrl(urlsCollection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	log.Println("Inserting")
	_, err := urlsCollection.InsertOne(ctx, bson.M{"url": "https://test.com"})

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Println("Test url already created before")
		} else {
			log.Println(err)
			log.Fatal("Error in inserting")
		}
	}

	defer cancel()
}

func createUniqueUrlIndex(urlsCollection *mongo.Collection) {
	_, err := urlsCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "url", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)

	if err != nil {
		log.Println(err)
		log.Fatal("Failed to create unique index")
	}
}
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH,OPTIONS,GET,PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	godotenv.Load(".env.local")
	MONGODB_URI, isMongoDbURISuppliedFromEnvFile := os.LookupEnv("MONGODB_URI")
	DATABASE_NAME, isDbNameSuppliedFromEnvFile := os.LookupEnv("DATABASE_NAME")
	DATABASE_COLLECTION_NAME, isCollectionNameSuppliedFromEnvFile := os.LookupEnv("DATABASE_COLLECTION_NAME")
	PORT, isPortSuppliedFromEnvFile := os.LookupEnv("PORT")

	reportBackendFs := flag.NewFlagSet("elasticpwn-bakcend", flag.ContinueOnError)
	mongodbURI := reportBackendFs.String("mongodbUri", "", "mongodb URI. Example: mongodb+srv://username:pw@somewhere.mongodb.net/default-collection-name?retryWrites=true&w=majority")
	databaseName := reportBackendFs.String("databaseName", "", "mongodb db name")
	databasecollectionName := reportBackendFs.String("databaseCollectionName", "", "mongodb collection name. Should be any one of: elasticsearch_reviewed_urls|elasticsearch_reviewed_urls_dev|kibana_reviewed_urls|kibana_reviewed_urls_dev")
	port := reportBackendFs.String("port", "9292", "port at which the server will run.")

	if !isMongoDbURISuppliedFromEnvFile || !isDbNameSuppliedFromEnvFile || !isCollectionNameSuppliedFromEnvFile || !isPortSuppliedFromEnvFile {
		if err := reportBackendFs.Parse(os.Args[1:]); err != nil {
			log.Fatal("At least one env var is nil. Check .env.local file. Otherwise, supply correct flags. Precedence: .env.local over flags")
		}
	}

	if reportBackendFs.Parsed() {
		if !isMongoDbURISuppliedFromEnvFile && *mongodbURI == "" {
			reportBackendFs.PrintDefaults()
			log.Fatal("MONGODB_URI was not found in either .env.local file or flag input")
		} else if !isDbNameSuppliedFromEnvFile && *databaseName == "" {
			reportBackendFs.PrintDefaults()
			log.Fatal("DATABASE_NAME was not found in either .env.local file or flag input")
		} else if !isCollectionNameSuppliedFromEnvFile && *databasecollectionName == "" {
			reportBackendFs.PrintDefaults()
			log.Fatal("DATABASE_COLLECTION_NAME was not found in either .env.local file or flag input")
		} else if !isPortSuppliedFromEnvFile && *port == "" {
			reportBackendFs.PrintDefaults()
			log.Fatal("DATABASE_COLLECTION_NAME was not found in either .env.local file or flag input")
		}
	}

	finalMongodbURI := func() string {
		if isMongoDbURISuppliedFromEnvFile {
			return MONGODB_URI
		} else {
			return *mongodbURI
		}
	}()
	finalDatabaseName := func() string {
		if isDbNameSuppliedFromEnvFile {
			return DATABASE_NAME
		} else {
			return *databaseName
		}
	}()
	finalCollectionName := func() string {
		if isCollectionNameSuppliedFromEnvFile {
			return DATABASE_COLLECTION_NAME
		} else {
			return *databasecollectionName
		}
	}()
	finalPort := func() string {
		if isPortSuppliedFromEnvFile {
			return PORT
		} else {
			return *port
		}
	}()
	if finalCollectionName != "elasticsearch_reviewed_urls" &&
		finalCollectionName != "elasticsearch_reviewed_urls_dev" &&
		finalCollectionName != "kibana_reviewed_urls" &&
		finalCollectionName != "kibana_reviewed_urls_dev" {
		reportBackendFs.PrintDefaults()
		log.Fatal("DATABASE_COLLECTION_NAME (-databaseCollectionName) should be any one of: elasticsearch_reviewed_urls|elasticsearch_reviewed_urls_dev|kibana_reviewed_urls|kibana_reviewed_urls_dev")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(finalMongodbURI))
	log.Println("Connected to database")
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	if err != nil {
		log.Fatal(err)
	}

	collection := mongoClient.Database(finalDatabaseName).Collection(finalCollectionName)
	insertTestUrl(collection)
	createUniqueUrlIndex(collection)

	r := gin.Default()
	r.Use(CORSMiddleware())
	r.GET("/ping", PingHandler)
	r.GET("/urls", GetUrlsHandlerGenerator(collection))
	r.POST("/urls", PostUrlsHandlerGenerator(collection))
	r.DELETE("/urls", DeleteUrlsHandlerGenerator(mongoClient, finalDatabaseName))

	serverHost := ""
	if runtime.GOOS == "windows" {
		serverHost = "localhost"
	} else {
		serverHost = "0.0.0.0"
	}
	r.Run(fmt.Sprintf("%v:%v", serverHost, finalPort))
}
