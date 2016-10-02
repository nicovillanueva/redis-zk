package main

import (
	"fmt"
	"gopkg.in/redis.v4"
	"os"
	"strings"
)

func GetZookeeperHosts() []string {
	if os.Getenv("ZK_HOSTS") == "" {
		return []string{"localhost:2181"}
	} else {
		return strings.Split(os.Getenv("ZK_HOSTS"), ",")
	}
}

func main() {
	fmt.Println("Starting")

	keeper := Keeper{GetZookeeperHosts()}
	masterChan := make(chan NodeReference)

	sentinel, err := keeper.GetRandomSentinel()
	if err != nil {
		panic(err)
	}

	wm := WatchOptions{
		Topic:     "+switch-master",
		HostIndex: 3,
		PortIndex: 4,
		RedisOpts: redis.Options{
			Addr:     sentinel.GetDottedAddr(),
			Password: "",
			DB:       0,
		},
	}

	go WatchTopic(wm, masterChan)

	go keeper.RunBehavior(masterChan, MasterUpdateBehavior)

	serve()
}
