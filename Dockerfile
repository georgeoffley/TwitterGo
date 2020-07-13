FROM golang:1.14.1
WORKDIR /home
ENV ENV GOPATH
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y git \
    go get -d -v github.com/joho/godotenv \
    go get -d -v github.com/gorilla/mux \
    go get -d -v github.com/amit-lulla/twitterapi
CMD ["go","run","main.go"]