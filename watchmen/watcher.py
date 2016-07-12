from redis import Redis
from threading import Thread, Event
import time

INTERVAL = 1

def get_subscriber(host='localhost', port=26379, topic='+switch-master'):
    r = Redis(host=host, port=port)
    ps = r.pubsub()
    ps.subscribe(topic)
    return ps


def watch_for_master():
    # +switch-master master1 127.0.0.1 6379 127.0.0.1 6380
    ps = get_subscriber(topic='+switch-master')
    while running.is_set():
        msg = ps.get_message()
        if msg is not None and type(msg.get('data')) is bytes:
            m = msg.get('data').decode('utf-8').split()
            print("Master name: {}\nMaster IP: {}\nMaster port: {}".format(m[0], m[3], m[4]))
        time.sleep(INTERVAL)
    print("shutdown event, closing")
    ps.close()


def watch_for_slaves():
    # +slave slave 127.0.0.1:6381 127.0.0.1 6381 @ master1 127.0.0.1 6380
    ps = get_subscriber(topic='+slave')
    while running.is_set():
        msg = ps.get_message()
        if msg is not None and type(msg.get('data')) is bytes:
            m = msg.get('data').decode('utf-8').split()
            print("New slave host: {}\nslave port: {}".format(m[2], m[3]))
        time.sleep(INTERVAL)
    print("shutdown event, closing")
    ps.close()

m = Thread(target=watch_for_master)
s = Thread(target=watch_for_slaves)
running = Event()
running.set()
m.start()
s.start()
try:
    while 1:
        time.sleep(0.1)
except:
    print("clearing event")
    running.clear()
    print("cleared")
    m.join()
    s.join()
    print("done")
