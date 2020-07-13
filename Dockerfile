FROM golang:1.14.1
WORKDIR /go/src/
ENV ENV GOPATH
COPY . .
CMD ["go","run","main.go"]