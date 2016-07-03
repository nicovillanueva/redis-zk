# Image: redis:3.2.1-alpine

from kazoo.client import KazooClient
import os

BASE_PATH="/redis"
MASTER_PATH="{}/master".format(BASE_PATH)
IDENTIFIER=os.getpid()

zk = KazooClient(hosts=os.environ.get("ZK_NODES") or '127.0.0.1:2181')
zk.start()

zk.delete(BASE_PATH, recursive=True)

zk.stop()
