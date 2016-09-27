/**
These functions are fired when "something" happens.
They are fired by the Keeper when it receives a NodeReference in a channel
*/

package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
)

type Behavior func(nr NodeReference, conn *zk.Conn) error

func MasterUpdateBehavior(nr NodeReference, conn *zk.Conn) error {
	fmt.Println("Got failover event. Setting new master: " + nr.GetDottedAddr())
	_, _, err := conn.Get("/redis/master")
	if err != nil {
		return fmt.Errorf("Can't get /redis/master node info")
	}
	_, err = conn.Set("/redis/master", []byte(nr.GetAddress()), -1)
	if err != nil {
		return fmt.Errorf("Could not set new master!")
	}
	return nil
}

/*func CleanSentinelsBehavior(nr NodeReference, conn *zk.Conn) error {
// TODO
}*/
