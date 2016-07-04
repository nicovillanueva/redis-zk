# ZK-coordinated Marathon-compatible Redis replication deployment

## Run
docker run -ti --rm --net host -e PORT0=6379 -e NET_IFACE='wlan0' redis-zk

## TODO
- Sentinel-aware ZK maintenance on failover (polling? sentinel watcher?)
- Coordinate Sentinels deployment
