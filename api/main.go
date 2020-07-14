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

// DB Connections

// TODO: Figure out how to get the DB connection up and how to populate it with collections etc
// API returns
func SearchPhrase(write http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	write.Header().Set("Content-Type", "text/html")
	write.WriteHeader(http.StatusOK)
	write.Write([]byte(params["searchInput"]))
}

func main() {
	fmt.Println("Hello from your Twitter container")

	////// Twitter Stuff
	api := CreateTwitterConn()

	// TODO: Set up Function for searching
	searchResult, _ := api.GetSearch("hip-hop", nil)

	for _, tweet := range searchResult.Statuses {
		fmt.Printf("UserName: %+v\n", tweet.User.ScreenName)
		fmt.Printf("TweetId: %+v\n", tweet.Id)
		fmt.Printf("Tweet Text: %+v\n", tweet.Text)
		fmt.Printf("Liked Count: %+v\n", tweet.FavoriteCount)
		fmt.Printf("Retweet Count: %+v\n", tweet.RetweetCount)
		fmt.Printf("Created At: %+v\n", tweet.CreatedAt)
	}

	// TODO: Put into function for multiple uses
	clientoptions := options.Client().ApplyURI("mongodb://icx-db-mongo:27017")
	client, err := mongo.Connect(context.TODO(), clientoptions)
	if err != nil {
		log.Fatal(err)
	}

	// Testing connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Connected to Mongo!")

	// TODO: Put the collection in it's own function since it needs to be reused for readng and writing
	collection := client.Database("icxSocial").Collection("icxSocial")
	Record := SocialRecord{45, "Someone", "This is a tweet", 4, 1, "07/13/2020 06:00:00 AM"}

	insert, err := collection.InsertOne(context.TODO(), Record)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted Single Doc: ", insert.InsertedID)

	/////// ROuter STuff
	fmt.Print("Starting Server...")

	router := mux.NewRouter()
	router.Queries("searchInput", "{searchInput}")
	router.HandleFunc("/searchphrase/{searchInput}", SearchPhrase)
	log.Fatal(http.ListenAndServe(":8080", router))
}
