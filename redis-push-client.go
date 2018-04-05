/*
 * Cyclick Development (Pete Washer)
 * This code based on: https://github.com/eclipse/paho.mqtt.golang/blob/master/cmd/sample/main.go
 * Additions: Unpack msgpack packed payloads, store the last message in redis.
 *
 * Original license:
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package main

import (
    "flag"
    "fmt"
    "os"
    "time"
    "encoding/json"
    Redis "github.com/go-redis/redis"
    MQTT "github.com/eclipse/paho.mqtt.golang"
    msgpack "github.com/vmihailenco/msgpack"
)

func RedisClient() (*Redis.Client) {
    client := Redis.NewClient(&Redis.Options{
        Addr:     "redis:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    return client
}

/*
Options:
[-help]                      Display help
[-a pub|sub]                 Action pub (publish) or sub (subscribe)
[-m <message>]               Payload to send
[-n <number>]                Number of messages to send or receive
[-q 0|1|2]                   Quality of Service
[-clean]                     CleanSession (true if -clean is present)
[-id <clientid>]             CliendID
[-user <user>]               User
[-password <password>]       Password
[-broker <uri>]              Broker URI
[-topic <topic>]             Topic
[-store <path>]              Store Directory
*/

type UnpackedMQTTEvent struct {
    TopicId string
    Payload string
}

type Item struct {
    Payload string
}

func main() {

    // Give time for the broker to wake up in the docker environment
    time.Sleep(10 * time.Second)

    // Check the redis client can talk to redis OK
    redis_client := RedisClient()
    err := redis_client.Set("hello", "world", (10 * time.Second)).Err()
    if err != nil {
        panic(err)
    }

    val, err := redis_client.Get("hello").Result()
    if err != nil {
        panic(err)
    }

    if val != "world" {
        panic(val)
    }

    fmt.Println("Successfully contacted redis")

    broker := flag.String("broker", "tcp://broker:1883", "The broker URI. ex: tcp://10.10.1.1:1883")

    topic := flag.String("topic", "my-topic", "The topic name to/from which to publish/subscribe")
    password := flag.String("password", "", "The password (optional)")
    user := flag.String("user", "", "The User (optional)")

    id := flag.String("id", "testgoid", "The ClientID (optional)")

    cleansess := flag.Bool("clean", false, "Set Clean Session (default false)")
    qos := flag.Int("qos", 0, "The Quality of Service 0,1,2 (default 0)")
    num := flag.Int("num", 1, "The number of messages to publish or subscribe (default 1)")
    payload := flag.String("message", "", "The message text to publish (default empty)")
    action := flag.String("action", "", "Action publish or subscribe (required)")
    store := flag.String("store", ":memory:", "The Store Directory (default use memory store)")
    flag.Parse()

    fmt.Printf("MQTT Connection Info:\n")
    fmt.Printf("\taction:    %s\n", *action)
    fmt.Printf("\tbroker:    %s\n", *broker)
    fmt.Printf("\tclientid:  %s\n", *id)
    fmt.Printf("\tuser:      %s\n", *user)
    fmt.Printf("\tpassword:  %s\n", *password)
    fmt.Printf("\ttopic:     %s\n", *topic)
    fmt.Printf("\tmessage:   %s\n", *payload)
    fmt.Printf("\tqos:       %d\n", *qos)
    fmt.Printf("\tcleansess: %v\n", *cleansess)
    fmt.Printf("\tnum:       %d\n", *num)
    fmt.Printf("\tstore:     %s\n", *store)

    opts := MQTT.NewClientOptions()
    opts.AddBroker(*broker)
    opts.SetClientID(*id)
    opts.SetUsername(*user)
    opts.SetPassword(*password)
    opts.SetCleanSession(*cleansess)
    if *store != ":memory:" {
        opts.SetStore(MQTT.NewFileStore(*store))
    }

    choke := make(chan [2]string)

    opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
        choke <- [2]string{msg.Topic(), string(msg.Payload())}
    })

    client := MQTT.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    if token := client.Subscribe(*topic, byte(*qos), nil); token.Wait() && token.Error() != nil {
        fmt.Println(token.Error())
        os.Exit(1)
    }

    var item Item
    for true {
        incoming := <-choke

        err = msgpack.Unmarshal([]byte(incoming[1]), &item)
        if err != nil {
            panic(err)
        }

        topic_id := string(incoming[0])
        fsevent := UnpackedMQTTEvent{topic_id, item.Payload}
        ev, err := json.Marshal(fsevent)

        if err != nil {
            panic(err)
        }

        redis_client.LPush("incoming-events", ev)
    }

    client.Disconnect(250)
    fmt.Println("Sample Subscriber Disconnected")
}