package EPUtils

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// This is a user defined method that returns mongo.Client,
// context.Context, context.CancelFunc and error.
// mongo.Client will be used for further database operation.
// context.Context will be used set deadlines for process.
// context.CancelFunc will be used to cancel context and
// resource associtated with it.

func connectToMongodb(url string) (*mongo.Client, context.Context,
	context.CancelFunc, error) {
	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	ctx, cancel := context.WithTimeout(context.Background(),
		10*time.Second)

	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	return client, ctx, cancel, err
}

// This is a user defined method that accepts
// mongo.Client and context.Context
// This method used to ping the mongoDB, return error if any.
func ping(client *mongo.Client, ctx context.Context) error {

	// mongo.Client has Ping to ping mongoDB, deadline of
	// the Ping method will be determined by cxt
	// Ping method return error if any occored, then
	// the error can be handled.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	EPLogger("Connected to MongoDB")
	return nil
}

func InitMongoConnnection(mongoUrl string) *mongo.Client {
	mongoClient, _, cancelConnectToMongodb, mongodbConnectErr := connectToMongodb(mongoUrl)
	if mongodbConnectErr != nil {
		panic(mongodbConnectErr)
	}
	defer cancelConnectToMongodb()
	pingContext, cancelPing := context.WithTimeout(context.Background(),
		10*time.Second)
	defer cancelPing()
	pingError := ping(mongoClient, pingContext)

	if pingError != nil {
		EPLogger("Failed to connect to mongodb. Check your mongodb config.")
		panic(pingError)
	}

	return mongoClient
}
