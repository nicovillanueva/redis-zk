# Image: redis:3.2.1-alpine

from kazoo.client import KazooClient
import os

BASE_PATH="/redis"
MASTER_PATH="{}/master".format(BASE_PATH)
IDENTIFIER=os.getpid()

zk = KazooClient(hosts='192.168.74.22:2181')
zk.start()

if not zk.exists(BASE_PATH):
    print("Nothing here")
else:
    print(zk.get("/redis/master"))
    print(zk.get_children("/redis/slaves"))

zk.stop()
