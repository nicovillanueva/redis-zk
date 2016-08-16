package main

import (
    "fmt"
    "gopkg.in/redis.v4"
    "github.com/samuel/go-zookeeper/zk"
    "time"
    "strings"
)

type nodeReference struct {
    Host string
    Port string
}

func (nr nodeReference) GetAddress() string {
    return nr.Host + " " + nr.Port
}

type WatchOptions struct {
    Topic string
    // In the caught message, where will the host/port be?
    HostIndex int
    PortIndex int
}

func watchTopic(watchOpts WatchOptions, redisOpts redis.Options, ch chan nodeReference) {
    c := redis.NewClient(&redisOpts)
    ps, err := c.Subscribe(watchOpts.Topic)
    if err != nil {
        panic(err)
    } else {
        fmt.Println("Connected to " + redisOpts.Addr)
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
        fmt.Println("Found event '%s' in topic '%s'", msg.Payload, msg.Channel)
        info := strings.Split(msg.Payload, " ")
        node := nodeReference{info[watchOpts.HostIndex], info[watchOpts.PortIndex]}
        ch <- node
    }
}

func RecordKeeper(ch chan nodeReference, b Behavior, zkh []string) {
    for {
        nr := <- ch
        b(nr, getZKConnection(zkh))
    }
    // REVIEW: Reuse connection?
}

// What to do
type Behavior func (nr nodeReference, conn *zk.Conn) bool

func AddSlaveBehavior(nr nodeReference, conn *zk.Conn) bool {
    defer conn.Close()
    slaves, _, err := conn.Children("/redis/slaves")
    if err != nil {
        fmt.Println("Could not get slaves list!")
        return false
    }
    nextIndex := len(slaves)
    _, err = conn.Create("/redis/slaves/" + string(nextIndex), []byte(nr.GetAddress()), 0, zk.WorldACL(zk.PermAll))
    if err != nil {
        fmt.Println("Could not create slave!")
        return false
    } else {
        fmt.Printf("Slave %s added\n", nr.GetAddress())
        return true
    }
}

func MasterUpdateBehavior(nr nodeReference, conn *zk.Conn) bool {
    defer conn.Close()
    _, err := conn.Set("/redis/master", []byte(nr.GetAddress()), 0)
    if err != nil {
        fmt.Println("Could not set new master!")
        panic(err)
    }
    return true
}

func setUpZookeeper(zkHosts []string) {
    conn := getZKConnection(zkHosts)
    defer conn.Close()
    fmt.Println("got zk conn")
    CreatePath(conn, "/redis")
    go CreatePath(conn, "/redis/master")
    go CreatePath(conn, "/redis/slaves")
    CreatePath(conn, "/redis/sentinels")
}

func getZKConnection(zkHosts []string) *zk.Conn {
    conn, _, err := zk.Connect(zkHosts, 10 * time.Second)
    if err != nil {
        fmt.Println("Could not connect to ZK!")
        panic(err)
    } else {
        fmt.Println("Connected to ZK")
    }
    return conn
}

func CreatePath(conn *zk.Conn, p string)  {
    // TODO: Make recursive
    if e, _, err := conn.Exists(p) ; ! e && err == nil {
        _, err := conn.Create(p, nil, 0, zk.WorldACL(zk.PermAll))
        if err != nil {
            fmt.Printf("Error creating path: %s\n", p)
            panic(err)
        }
    } else if err != nil {
        fmt.Println("Error checking existance of path %s\n", p)
        panic(err)
    } else {
        fmt.Printf("Path %s already exists\n", p)
    }
}

func main() {
    fmt.Println("Starting")

    rOpts := redis.Options {
        Addr: "192.168.0.20:19107",
        Password: "",
        DB: 0,
    }

    wm := WatchOptions {
        Topic: "+switch-master",
        HostIndex: 3,
        PortIndex: 4,
    }
    ws := WatchOptions {
        Topic: "+slave",
        HostIndex: 2,
        PortIndex: 3,
    }

    masterChan := make(chan nodeReference)
    slaveChan := make(chan nodeReference)
    done := make(chan bool)
    zkh := []string{"localhost:2181"}
    setUpZookeeper(zkh)
    go watchTopic(wm, rOpts, masterChan)
    go watchTopic(ws, rOpts, slaveChan)

    go RecordKeeper(masterChan, MasterUpdateBehavior, zkh)
    go RecordKeeper(slaveChan, AddSlaveBehavior, zkh)

    // Just block
    if <- done {
        fmt.Println("done!")
    }
}
