/**
The Keeper fires a Behavior once it gets a NodeReference via a channel
*/

package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"math/rand"
	"strings"
	"time"
)

type Keeper struct {
	ZkHosts []string
}

func (k *Keeper) RunBehavior(ch chan NodeReference, b Behavior) {
	c := k.getZKConnection()
	for {
		nr := <-ch
		b(nr, c)
	}
	c.Close()
}

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
/*
func (k *Keeper) GetAllSentinels() ([]NodeReference, error) {
    conn := k.getZKConnection()
    defer conn.Close()
    sentinels := make([]NodeReference, 0)
    delta := 0
    for {
        child, _, err := conn.Children("/redis/sentinels")
        if err != nil {
            return sentinels, fmt.Errorf("Could not get sentinels list!")
        }
        if len(child) == 0 {
            if delta == 30 {
                return sentinels, fmt.Errorf("No sentinels found in 30 seconds. Aborting.")
            }
            fmt.Println("No sentinels found yet. Waiting.")
            time.Sleep(1 * time.Second)
            delta += 1
        } else {
            for cindex := 0; cindex < len(child); cindex++ {
                d, _, err := conn.Get("/redis/sentinels/" + string(cindex))
                if err != nil {
                    return sentinels, fmt.Errorf("Could not get sentinels!")
                }
                s := strings.Split(string(d), " ")
                sentinels = append(sentinels, NodeReference{s[0], s[1]})
            }
            return sentinels, nil
        }
    }
}
*/
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
