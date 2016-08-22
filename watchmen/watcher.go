package main

import (
    "os"
    "fmt"
    "gopkg.in/redis.v4"
    "github.com/samuel/go-zookeeper/zk"
    "time"
    "strings"
    "math/rand"
)

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
        fmt.Printf("Found event '%s' in topic '%s'\n", msg.Payload, msg.Channel)
        info := strings.Split(msg.Payload, " ")
        node := nodeReference{info[watchOpts.HostIndex], info[watchOpts.PortIndex]}
        ch <- node
    }
}

func RecordKeeper(ch chan nodeReference, b Behavior, zkh []string) {
    c := getZKConnection(zkh)
    for {
        nr := <- ch
        b(nr, c)
    }
    c.Close()
}


func getRandomSentinel(conn *zk.Conn) (nodeReference, error) {
    delta := 0
    for {
        //child, _, err := conn.Children("/redis/roles/sentinels")
        child, _, err := conn.Children("/redis/sentinels")
        if err != nil {
            return nodeReference{"", ""}, fmt.Errorf("Could not get sentinels list!")
        }
        if len(child) == 0 {
            if delta == 30 {
                return nodeReference{"", ""}, fmt.Errorf("No sentinels found in 30 seconds. Aborting.")
            }
            fmt.Println("No sentinels found yet. Waiting.")
            time.Sleep(1 * time.Second)
            delta += 1
        } else {
            c := child[rand.Intn(len(child))]
            // d, _, err := conn.Get("/redis/roles/sentinels/" + c)
            d, _, err := conn.Get("/redis/sentinels/" + c)
            if err != nil {
                return nodeReference{"", ""}, fmt.Errorf("Could not get sentinels!")
            }
            s := strings.Split(string(d), " ")
            fmt.Println("Using Sentinel: " + string(d))
            return nodeReference{s[0], s[1]}, nil
        }
    }
}

/*func findMembers(zkh []string) {
    c := getZKConnection(zkh)
    defer c.Close()
    child, _, err := c.Children("/redis/members")
    if err != nil {
        // FIXME
        panic(err)
    }
    for _, ch := range child {
        n, _, err := c.Get("/redis/members/" + ch)
        if err != nil {
            // FIXME
            panic(err)
        }
        fmt.Printf("%s -> %s\n", ch, string(n))
    }
}*/

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

/*func CreatePath(conn *zk.Conn, p string)  {
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
}*/

func main() {
    fmt.Println("Starting")

    /*rOpts := redis.Options {
        Addr: "192.168.0.20:19107",
        Password: "",
        DB: 0,
    }*/
    wm := WatchOptions {
        Topic: "+switch-master",
        HostIndex: 3,
        PortIndex: 4,
    }
    /*ws := WatchOptions {
        Topic: "+sdown",
        HostIndex: 2,
        PortIndex: 3,
    }*/

    var zkh []string
    if os.Getenv("ZK_HOSTS") == "" {
        zkh = []string{"localhost:2181"}
    } else {
        zkh = strings.Split(os.Getenv("ZK_HOSTS"), ",")
    }

    masterChan := make(chan nodeReference)
    //slaveChan := make(chan nodeReference)
    done := make(chan bool)

    zkc := getZKConnection(zkh)
    defer zkc.Close()
    sentinel, err := getRandomSentinel(zkc)
    if err != nil {
        panic(err)
    }

    rOpts := redis.Options {
        Addr: sentinel.GetDottedAddr(),
        Password: "",
        DB: 0,
    }

    //findMembers(zkh)

    //setUpZookeeper(zkh)
    go watchTopic(wm, rOpts, masterChan)
    //go watchTopic(ws, rOpts, slaveChan)

    go RecordKeeper(masterChan, MasterUpdateBehavior, zkh)
    //go RecordKeeper(slaveChan, CleanSlaveBehavior, zkh)

    // Just block
    if <- done {
        fmt.Println("done!")
    }
}
