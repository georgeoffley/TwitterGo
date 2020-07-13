// New app script. Will use for creating own API calls
package main

import (
	"fmt"
	"log"
	"os"

	//"github.com/gorilla/mux"
	"github.com/amit-lulla/twitterapi"

	"github.com/joho/godotenv"
)

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

func main() {
	fmt.Println("Hello from your Twitter container")

	api := CreateTwitterConn()

	searchResult, _ := api.GetSearch("golang", nil)
	for _, tweet := range searchResult.Statuses {
		fmt.Printf("UserName: %+v\n", tweet.User.ScreenName)
		fmt.Printf("TweetId: %+v\n", tweet.Id)
		fmt.Printf("Tweet Text: %+v\n", tweet.Text)
	}
}
