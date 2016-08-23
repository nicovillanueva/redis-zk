# Redis + ZK
(Pretencious name pending)
### DC/OS-compatible, highly available Redis replica "cluster"

## Purpose
Note: I'll refer to a group of Redis in a replica scheme as a "cluster". It's not correct but you are not my boss, probably.  
Given the nature of DC/OS (and Docker, and Mesos), it's quite unpredictable on which host & port a service will come up. You will usually create an application in Marathon, and -unless you set some constraints- it will pop up in any host, under any port.  
Given this, creating a Redis replica cluster "out of the box", it's near impossible, as you can never know what SLAVEOF to set.

## Requirements
- Apache Zookeeper
- Marathon
- Docker
- Or simply: DC/OS

## Architecture
The Redis and Sentinel components are simply Python wrapper scripts. The Watcher component is a Go application. All components require the environment variable `ZK_HOSTS`.

### Redis
Upon starting up, the script queries Zookeeper to find out which is the current Redis master. If it doesn't find one, it'll set itself as master. If it does indeed find one, it'll launch as a slave to it. This is done using ZK Locks, so it -in theory- is safe to throw many replicas at once and only one will stick.

### Sentinels
Before starting the Sentinel up, the script checks if there are Redis masters running. If it finds none, it will wait for a while, querying every second. Once it finds a master, it will start monitoring it. Given the nature of Sentinels, it auto-discovers it's slaves, and will perform failovers and such as needed.  
As soon as it finds a master to connect to, it will also register itself into Zookeeper, identified by an UUID.

### Watchers
The watcher's job is to keep Zookeeper up to date when a failover occurs. It queries Zookeeper for a list of current sentinels, and chooses one at random. Then, using the pub/sub protocol, it watches for changes in the masters.  
Once a failover occurs, it picks up the new master's IP and port, and it updates it in Zookeeper.

It's extensible for monitoring more events and attaching behaviors to them.

## Environment variables/Config
### Redis
- `ZK_HOSTS`: Comma-separated Zookeeper hosts. Ex: "zk1:2181,zk2:2222,zk3:8888". Defaults to "localhost:2181"
- `NET_IFACE`: Interface to bind to. Keep in mind that all hosts will need to have this interface available. Defaults to "eth0"
- `BIND_IP`: Directly bind to a specific IP. Overrides `NET_IFACE`. It has no default value.
- `PORT0`: On what port will the Redis server run. When running on Marathon, this gets auto-populated with the mapped port. Usually, you don't need to set this. If it does not find the variable, it defaults to 6379 (for example, when running locally, without Marathon)

### Sentinel
- `ZK_HOSTS`: Comma-separated Zookeeper hosts. Ex: "zk1:2181,zk2:2222,zk3:8888". Defaults to "localhost:2181"
- `SENTINEL_QUORUM`: Quorum to use when deciding on a new master. Defaults to 2.
- `NET_IFACE`: Interface to bind to. Keep in mind that all hosts will need to have this interface available. Defaults to "eth0"
- `BIND_IP`: Directly bind to a specific IP. Overrides NET_IFACE. It has no default value.
- `PORT0`: On what port will the Redis sentinel run. When running on Marathon, this gets auto-populated with the mapped port. Usually, you don't need to set this. If it does not find the variable, it defaults to 26379 (for example, when running locally, without Marathon)

You may configure Redis and the Sentinel themselves by modifying the `redis.conf` and `sentinel.conf` files

### Watcher
- `ZK_HOSTS`: Comma-separated Zookeeper hosts. Ex: "zk1:2181,zk2:2222,zk3:8888". Defaults to "localhost:2181"

## Implementation
First and foremost, you need to use a Redis client which supports Sentinels. Most libraries do, such as Jedis. This library will connect to the Sentinel, and from there, connect to the Master.  
In order to get the host/port of a Sentinel, you should first query Zookeeper. On the path `/redis/sentinels` are listed all of the current Sentinels. You could for example list them all, and pick one at random. Then do a get on that zknode and you'll get the -space-separated- host and port of a Sentinel.

## Paths in ZK
- `/redis`: Root, just that. It's empty
- `/redis/master`: Current master. Do a get on it, and you'll get the host and port of the current master.
- `/redis/sentinels`: List of current sentinels. For example, `/redis/sentinels/some-random-uuid`. Do a get on any of the elements of the list, and you'll get the host/port of a sentinel.

## TODO
- Wrappers: print -> logger
- Watcher: Watch -sentinel (sentinel cleanup)
- Readme
- Sentinel: Configure timeout when checking for masters
