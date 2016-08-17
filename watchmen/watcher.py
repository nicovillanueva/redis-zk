#
## DEPRECATED
#
from redis import Redis
from kazoo.client import KazooClient
from kazoo.exceptions import NoNodeError
from threading import Thread, Event
import time, os, random

zk_hosts = os.environ.get("ZK_HOSTS") or '127.0.0.1:2181'
base_path = '/redis'
master_path = '{}/master'.format(base_path)
slaves_path = '{}/slaves'.format(base_path)
sentinels_path = '{}/sentinels'.format(base_path)

interface = os.environ.get("NET_IFACE") or 'eth0'
port = os.environ.get("PORT0") or "6379"

zk = KazooClient(hosts=zk_hosts)

interval = 1

def find_sentinel():
    zk.start()
    sent = random.choice(zk.get_children(sentinels_path))
    sent = sentinels_path + '/' + sent
    #with zk.Lock(sent) as lock:
    v, s = zk.get(sent)
    v = v.decode('utf-8')
    print('Got sentinel in {}'.format(v))
    h, p = v.split()
    zk.stop()
    return h, p


def get_subscription(topic='+switch-master'):
    h, p = find_sentinel()
    r = Redis(host=h, port=p)
    ps = r.pubsub()
    ps.subscribe(topic)
    return ps


def watch_for_master():
    # +switch-master master1 127.0.0.1 6379 127.0.0.1 6380
    ps = get_subscription(topic='+switch-master')
    while running.is_set():
        msg = ps.get_message()
        if msg is not None and type(msg.get('data')) is bytes:
            m = msg.get('data').decode('utf-8').split()
            print("New master: {}!\nHere: {} {}".format(m[0], m[3], m[4]))
            zk.start()
            zk.ensure_path(master_path)
            zk.set(master_path, bytes('{} {}'.format(m[3], m[4]), encoding="UTF-8"))
            zk.stop()
        time.sleep(interval)
    ps.close()


def watch_for_slaves():
    # +slave slave 127.0.0.1:6381 127.0.0.1 6381 @ master1 127.0.0.1 6380
    ps = get_subscription(topic='+slave')
    while running.is_set():
        msg = ps.get_message()
        if msg is not None and type(msg.get('data')) is bytes:
            m = msg.get('data').decode('utf-8').split()
            print("New slave!\nHere: {} {}".format(m[2], m[3]))
            #
            # TODO:
            # - Update current slaves
            # - Clean dead slaves
            # - Test with multiple watchers
            #
        time.sleep(interval)
    ps.close()

m = Thread(target=watch_for_master)
s = Thread(target=watch_for_slaves)
print('Watching...')
running = Event()
running.set()

try:
    m.start()
    s.start()
    while 1:
        time.sleep(0.1)
except:
    print("Shutting down")
    running.clear()
    m.join()
    s.join()
    print("Done")
