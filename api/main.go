// New app script. Will use for creating own API calls
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/amit-lulla/twitterapi"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Using one connection to DB
var client = *CreateDBCon()

// Saving todo since the context can be vague
var ctx = context.TODO()

// Constants for Collections and DB
const dbname = "icxSocial"
const collectionName = "socialTweets"

///// Custom Data Types

// APICred Struct for storing credentials
type APICred struct {
	APIKey            string
	APISecret         string
	AccessToken       string
	AccessTokenSecret string
}

// Also hold json and bson formatting for easier working with json
type SocialRecord struct {
	TweetId  int64  `json:"tweetid" bson:"tweetid"`
	UserName string `json:"username" bson:"username"`
	Tweet    string `json:"tweet" bson:"tweet"`
	Likes    int    `json:"likes" bson:"likes"`
	Retweets int    `json:"retweets" bson:"retweets"`
	Created  string `json:"created" bson:"created"`
}

/////// Twitter Functions
// TODO: Make concurrent
// Load Env file and fill out credentials for API
func LoadEnv() (env APICred) {
	// TODO: Make this concurrent
	err := godotenv.Load()

	// Establish credentials for accessing API
	env = APICred{
		os.Getenv("API_KEY"),
		os.Getenv("API_SECRET"),
		os.Getenv("ACCESS_TOKEN"),
		os.Getenv("ACCESS_TOKEN_SECRET")}
	// Credential error checking
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return env
}

// Already running concurrently in lib code
// Set up and create twitter API
func CreateTwitterConn() (api *twitterapi.TwitterApi) {

	// Grab Credentials and load into struct
	credentials := LoadEnv()

	// Connect to Twitter API
	twitterapi.SetConsumerKey(credentials.APIKey)
	twitterapi.SetConsumerSecret(credentials.APISecret)
	api = twitterapi.NewTwitterApi(credentials.AccessToken, credentials.AccessTokenSecret)

	return api
}

// Create Search
func CreateTwitSearch(api *twitterapi.TwitterApi, query string) (searchResult twitterapi.SearchResponse) {
	searchResult, _ = api.GetSearch(query, nil)

	return searchResult
}

//////// DB Functions

// DB Connection
func CreateDBCon() (client *mongo.Client) {
	// TODO: Put into function for multiple uses
	clientoptions := options.Client().ApplyURI("mongodb://icx-db-mongo:27017")
	client, err := mongo.Connect(ctx, clientoptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Connected to Mongo!\n")

	return client
}

// Create Collection for sending data to
func CollectionItem(client *mongo.Client, dbName string, collectionNam string) (collection *mongo.Collection) {
	collection = client.Database(dbName).Collection(collectionNam)

	return collection
}

///// API Stuff

// Search all docs in collection
func ReturnAllDocs(client *mongo.Client, collection *mongo.Collection) (results []bson.D) {

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var records []bson.D
	/*
		In production this would be cursor.Next() so that we're returning
		the data in batches and not all at once. A production system would have millions
		of records (documents) to return
	*/
	if err = cursor.All(ctx, &records); err != nil {
		log.Fatal(err)
	}
	return records
}

// Takes the collection and sorts by the indicated category and returns the top result
func SearchMostPopularBySubject(category string) (mostPopularRecord SocialRecord) {
	grabTweetCollection := CollectionItem(&client, dbname, collectionName)

	opts := options.FindOne().SetSort(bson.D{{category, -1}})

	err := grabTweetCollection.FindOne(ctx, bson.D{}, opts).Decode(&mostPopularRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return
			log.Fatal(err)
		}
		log.Fatal("Error")
	}

	return mostPopularRecord
}

///// API Calls

// Search For most liked tweets and return to endpoint
func SearchMostLikedTweet(write http.ResponseWriter, req *http.Request) {
	write.Header().Set("Content-Type", "application/json")
	mostLiked := SearchMostPopularBySubject("likes")
	json.NewEncoder(write).Encode(mostLiked)
}

// Search Most retweeted tweet
func SearchMostRtTweet(write http.ResponseWriter, req *http.Request) {
	write.Header().Set("Content-Type", "application/json")
	mostRt := SearchMostPopularBySubject("retweets")
	json.NewEncoder(write).Encode(mostRt)
}

// Search for all tweets
func SearchAll(write http.ResponseWriter, req *http.Request) {
	write.Header().Set("Content-Type", "application/json")
	grabTweetCollection := CollectionItem(&client, dbname, collectionName)
	alldocs := ReturnAllDocs(&client, grabTweetCollection)
	json.NewEncoder(write).Encode(alldocs)
}

//

func main() {
	fmt.Print("Hello from your Twitter container\n")

	/////// DB Stuff
	/*
		Reuse connection pool below
		so that we can do not have to keep opening DB connections
	*/
	dbclient := &client
	tweetcollection := CollectionItem(dbclient, dbname, collectionName)

	////// Twitter Stuff
	api := CreateTwitterConn()
	searchResult := CreateTwitSearch(api, "hip-hop")

	// Initialize Bulk Record Collection and Record data type
	var BulkRecords []interface{}
	tempSocialRecord := SocialRecord{}

	// Iterate through tweets and add to single interface
	for _, tweet := range searchResult.Statuses {
		tempSocialRecord.TweetId = tweet.Id
		tempSocialRecord.UserName = tweet.User.ScreenName
		tempSocialRecord.Tweet = tweet.Text
		tempSocialRecord.Likes = tweet.FavoriteCount
		tempSocialRecord.Retweets = tweet.RetweetCount
		tempSocialRecord.Created = tweet.CreatedAt

		// Append interface on each loop
		BulkRecords = append(BulkRecords, tempSocialRecord)
	}

	// Take Populated Interface and Insert Records into DB
	insert, err := tweetcollection.InsertMany(ctx, BulkRecords)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted Many Docs: %+v\n", insert.InsertedIDs)

	// Testing
	//fmt.Println(ReturnAllDocs(dbclient, tweetcollection))

	/////// ROuter STuff
	fmt.Print("Starting API Server...\n")

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/mostliked", SearchMostLikedTweet)
	router.HandleFunc("/api/v1/mostrt", SearchMostRtTweet)
	router.HandleFunc("/api/v1/alltweets", SearchAll)

	log.Fatal(http.ListenAndServe(":8080", router))
}
