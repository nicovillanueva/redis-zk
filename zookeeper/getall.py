# Image: redis:3.2.1-alpine

from kazoo.client import KazooClient
import os

BASE_PATH="/redis"
MASTER_PATH="{}/master".format(BASE_PATH)
IDENTIFIER=os.getpid()

zk = KazooClient(hosts=os.environ.get("ZK_NODES") or '127.0.0.1:2181')
zk.start()

if not zk.exists(BASE_PATH):
    print("Nothing here")
else:
    print(zk.get("/redis/master"))
    print(zk.get_children("/redis/slaves"))

zk.stop()
