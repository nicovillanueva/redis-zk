# Image: redis:3.2.1-alpine

from kazoo.client import KazooClient
import os

BASE_PATH="/redis"
MASTER_PATH="{}/master".format(BASE_PATH)
IDENTIFIER=os.getpid()

zk = KazooClient(hosts='192.168.74.22:2181')
zk.start()

zk.delete(BASE_PATH, recursive=True)

zk.stop()
