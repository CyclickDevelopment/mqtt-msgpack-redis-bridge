FROM golang:alpine3.7

RUN apk add --no-cache git
WORKDIR /go/src/app
RUN go get -u github.com/go-redis/redis
RUN go get github.com/eclipse/paho.mqtt.golang
RUN go get -u github.com/vmihailenco/msgpack

COPY . .
RUN go clean && go build
CMD ["./app"]