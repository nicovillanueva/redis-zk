package main

import (
	"fmt"
	"gopkg.in/redis.v4"
	"strings"
)

// WatchTopic waits for messages in a given NodeReference channel
// Once it gets a NR, sends it through the channel, the Keeper receives it
// and fires a Behavior (configured within the WatchOptions)
func WatchTopic(watchOpts WatchOptions, ch chan NodeReference) {
	c := redis.NewClient(&watchOpts.RedisOpts)
	ps, err := c.Subscribe(watchOpts.Topic)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Connected to " + watchOpts.RedisOpts.Addr)
	}
	defer ps.Close()

	for {
		fmt.Println("Watching topic: " + watchOpts.Topic)
		msg, err := ps.ReceiveMessage()
		if err != nil {
			fmt.Println("Error when receiving message from topic " + watchOpts.Topic)
			fmt.Println(err)
			return
		}
		fmt.Printf("Found event '%s' in topic '%s'\n", msg.Payload, msg.Channel)
		info := strings.Split(msg.Payload, " ")
		node := NodeReference{info[watchOpts.HostIndex], info[watchOpts.PortIndex]}
		ch <- node
	}
}
