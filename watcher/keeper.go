package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"math/rand"
	"strings"
	"time"
)

// Keeper fires a Behavior once it gets a NodeReference via a channel
type Keeper struct {
	ZkHosts []string
}

// RunBehavior blocks indefinitely waiting for events on a given channel
// Once it gets something, it opens a connection to ZK, and fires a behavior
// Afterwards, it closes it, preventing leaks or long-lived connections
func (k *Keeper) RunBehavior(ch chan NodeReference, b Behavior) {
	for {
		nr := <-ch
		c := k.getZKConnection()
		b(nr, c)
		c.Close()
	}
}

// GetRandomSentinel is a function that is proper of the Keeper, and not implemented as a behavior
// Looks into ZK, and returns a random Sentinel's ip/port
func (k *Keeper) GetRandomSentinel() (NodeReference, error) {
	conn := k.getZKConnection()
	defer conn.Close()
	delta := 0
	for {
		child, _, err := conn.Children("/redis/sentinels")
		if err != nil {
			return NodeReference{"", ""}, fmt.Errorf("Could not get sentinels list!")
		}
		if len(child) == 0 {
			if delta == 30 {
				return NodeReference{"", ""}, fmt.Errorf("No sentinels found in 30 seconds. Aborting.")
			}
			fmt.Println("No sentinels found yet. Waiting.")
			time.Sleep(1 * time.Second)
			delta += 1
		} else {
			c := child[rand.Intn(len(child))]
			d, _, err := conn.Get("/redis/sentinels/" + c)
			if err != nil {
				return NodeReference{"", ""}, fmt.Errorf("Could not get sentinels!")
			}
			s := strings.Split(string(d), " ")
			fmt.Println("Using Sentinel: " + string(d))
			return NodeReference{s[0], s[1]}, nil
		}
	}
}

// GetAllSentinels, similarly to GetRandomSentinel, is inherently linked to the Keeper
// Queries ZK, and gets a comma-separated list of all sentinels
// TODO: Code duplication sucks
func (k *Keeper) GetAllSentinels() ([]NodeReference, error) {
	conn := k.getZKConnection()
	defer conn.Close()
	sentinels := make([]NodeReference, 0)
	delta := 0
	for {
		child, _, err := conn.Children("/redis/sentinels")
		if err != nil {
			return sentinels, fmt.Errorf("Could not get sentinels list!\n%s\n", err.Error())
		}
		if len(child) == 0 {
			if delta == 30 {
				return sentinels, fmt.Errorf("No sentinels found in 30 seconds. Aborting.")
			}
			fmt.Println("No sentinels found yet. Waiting.")
			time.Sleep(1 * time.Second)
			delta += 1
		} else {
			for _, c := range child {
				d, _, err := conn.Get("/redis/sentinels/" + string(c))
				if err != nil {
					return sentinels, fmt.Errorf("Could not get sentinels!\n%s\n", err.Error())
				}
				s := strings.Split(string(d), " ")
				sentinels = append(sentinels, NodeReference{s[0], s[1]})
			}
			return sentinels, nil
		}
	}
}

func (k *Keeper) getZKConnection() *zk.Conn {
	conn, _, err := zk.Connect(k.ZkHosts, 10*time.Second)
	if err != nil {
		fmt.Println("Could not connect to ZK!")
		panic(err)
	} else {
		fmt.Println("Connected to ZK")
	}
	return conn
}
