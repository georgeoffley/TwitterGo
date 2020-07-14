// New app script. Will use for creating own API calls
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/amit-lulla/twitterapi"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// APICred Struct for storing credentials
type APICred struct {
	APIKey            string
	APISecret         string
	AccessToken       string
	AccessTokenSecret string
}

type SocialRecord struct {
	TweetId  int64
	UserName string
	Tweet    string
	Likes    int
	Retweets int
	Created  string
}

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

// DB Stuff

// DB Connection
func CreateDBCon() (client *mongo.Client) {
	// TODO: Put into function for multiple uses
	clientoptions := options.Client().ApplyURI("mongodb://icx-db-mongo:27017")
	client, err := mongo.Connect(context.TODO(), clientoptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
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

// API returns
func SearchPhrase(write http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	write.Header().Set("Content-Type", "text/html")
	write.WriteHeader(http.StatusOK)
	write.Write([]byte(params["searchInput"]))
}

func main() {
	fmt.Print("Hello from your Twitter container\n")

	/////// DB Stuff
	client := CreateDBCon()
	collection := CollectionItem(client, "icxSocial", "icxSocial")

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

		BulkRecords = append(BulkRecords, tempSocialRecord)
	}

	// Take Populated Interface and Insert Records into DB
	insert, err := collection.InsertMany(context.TODO(), BulkRecords)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted Many Docs: %+v\n", insert.InsertedIDs)

	/////// ROuter STuff
	fmt.Print("Starting Server...\n")

	router := mux.NewRouter()
	router.Queries("searchInput", "{searchInput}")
	router.HandleFunc("/searchphrase/{searchInput}", SearchPhrase)
	log.Fatal(http.ListenAndServe(":8080", router))
}
