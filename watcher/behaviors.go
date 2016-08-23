package main

import (
    "fmt"
    "github.com/samuel/go-zookeeper/zk"
)

type Behavior func (nr nodeReference, conn *zk.Conn) error

/*
update master - done
clean dead sentinels
*/

/*func AddSlaveBehavior(nr nodeReference, conn *zk.Conn) error {
    //defer conn.Close()
    slaves, _, err := conn.Children("/redis/roles/slaves")
    if err != nil {
        return fmt.Errorf("Could not get slaves list!")
    }
    nextIndex := len(slaves)
    _, err = conn.Create("/redis/roles/slaves/" + string(nextIndex), []byte(nr.GetAddress()), 0, zk.WorldACL(zk.PermAll))
    if err != nil {
        return fmt.Errorf("Could not create slave!")
    } else {
        fmt.Printf("Slave %s added\n", nr.GetAddress())
        return nil
    }
}*/

func MasterUpdateBehavior(nr nodeReference, conn *zk.Conn) error {
    //defer conn.Close()
    fmt.Println("Got failover event. Setting new master: " + nr.GetDottedAddr())
    //_, _, err := conn.Get("/redis/roles/master")
    _, _, err := conn.Get("/redis/master")
    if err != nil {
        //fmt.Println("Can't get /redis/roles/master node info")
        return fmt.Errorf("Can't get /redis/master node info")
    }
    //var vers int32 = 1000
    //fmt.Println("Node version: " + string(stat.Version))  // Que poronga de tipo es Version?
    //_, err = conn.Set("/redis/roles/master", []byte(nr.GetAddress()), -1)
    _, err = conn.Set("/redis/master", []byte(nr.GetAddress()), -1)
    if err != nil {
        return fmt.Errorf("Could not set new master!")
    }
    return nil
}

/*func CleanSentinelsBehavior(nr nodeReference, conn *zk.Conn) error {

}*/
