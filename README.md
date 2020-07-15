# ICX Social App
###### Not for use in production

## Running Containers

Running these assumes the user has Docker and Docker Compose Installed. To install, unzip the provided zip into a new directory. Then open terminal to the newly unzipped directory and run the following command:

```
docker-compose up --build
```

This should build the images linked to the docker-compose file and start them up. Developed on Windows 10 version 1903 and tested on MacOS 10.15.05

## File System

The file system layout should look as such:

icx_social_app
- api
    - .env
    - DockerFile
    - go.mod
    - go.sum
    - main.go
- web
    - Dockerfile
    - index.html
- .env
- .gitignore
- docker-compose.yml
- README.md

## Package Management

After some experimenting I discovered using Go Modules for external package manegement. For a large scale project having these libraries more local would be more practical.

## Enviornment Files

Enviornment files are included to keep them anonymous for sites like github and to take advantage of write once and read many style of storing confidential constants.

## System Layout

The system is laid out in three images. The MongoDB instance, the api instance, and the web end point instance. The api is where all the logic lives. The api is dependant on the DB running as DB operations are conducted on startup and whenever the API is hit. 

## Future Update Ideas

- When looking into all the records the amount of returned documents is relatively low. As this transitions into a large scale analytics program we are talking about millions and millions of documents. As a result we need to parcel out chunks of the data when searching for all records using cursor.Next(). For this assessment it makes more sense to just serve all the records using the cursor.All() function in the mongo driver.

- Loading metric data for tweets. We can calculate engagement rates using the engagements / the impressions. However this looks as if it is buried deeper in the API and I found it too late

- A tweet filter service needs to implemented. This includes not including poles, blank tweets, images, videos which might not be related to said subject

- An image and or video recognition system which can add weight to the popularity of a subject depending on the video

- A hashtag filter which can take all the hashtags via the api and store them with the tweets

- Possibly restructure the app so that the containers start, you send a query to an api function which will search for tweets. This would include making note of the subject and storing that with the records. This would also require a new way to change the api functions to only use the current running subject. Possibly another table with current subject