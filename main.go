// New app script. Will use for creating own API calls
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/amit-lulla/twitterapi"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type server struct{}

// APICred Struct for storing credentials
type APICred struct {
	APIKey            string
	APISecret         string
	AccessToken       string
	AccessTokenSecret string
}

// Make a custom data type for the returned tweets for easier handling
type SavedTweets struct {
	Id   int64
	User string
	Text string
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

// TODO: Figure out how to get the DB connection up and how to populate it with collections etc
// API returns
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// MongoDB connection
	fmt.Print("Connecting to DB...")

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to DB")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Connected to DB"}`))
}

func main() {
	fmt.Println("Hello from your Twitter container")

	api := CreateTwitterConn()

	searchResult, _ := api.GetSearch("golang", nil)
	for _, tweet := range searchResult.Statuses {
		fmt.Printf("UserName: %+v\n", tweet.User.ScreenName)
		fmt.Printf("TweetId: %+v\n", tweet.Id)
		fmt.Printf("Tweet Text: %+v\n", tweet.Text)
	}

	// Router Service Creation
	fmt.Print("Starting Server...")
	serv := &server{}
	http.Handle("/", serv)
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Print("Server Started!")
}
