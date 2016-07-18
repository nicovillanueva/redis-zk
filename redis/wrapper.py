from kazoo.client import KazooClient
import os, shlex, subprocess, socket, fcntl, struct, uuid

zk_hosts = os.environ.get("ZK_HOSTS") or '127.0.0.1:2181'
base_path = "/redis"
master_path = "{}/master".format(base_path)
slaves_path = "{}/slaves".format(base_path)

interface = os.environ.get("NET_IFACE") or 'eth0'
port = os.environ.get("PORT0") or "6379"
identifier = str(uuid.uuid4())
is_master = False

zk = KazooClient(hosts=zk_hosts)

def get_master_cmd():
    c = "redis-server /etc/redis/redis.conf --port {}".format(port)
    return shlex.split(c)

def get_slave_cmd(master):
    c = "redis-server /etc/redis/redis.conf --port {} --slaveof {}".format(port, master)
    return shlex.split(c)


def get_ip_address(ifname):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    return socket.inet_ntoa(fcntl.ioctl(
        s.fileno(),
        0x8915,  # SIOCGIFADDR
        struct.pack('256s', bytes(ifname, 'utf-8'))
    )[20:24])



zk.start()

mpath = zk.ensure_path(master_path)
my_path = ""
if type(mpath) is str:
    with zk.Lock(master_path, "{}-master".format(identifier)) as lock:
        print("Setting {} as master".format(identifier))
        s = "{} {}".format(get_ip_address(interface), port)
        zk.set(master_path, bytes(s, encoding="UTF-8"))
        is_master = True
        my_path = master_path
        cmd = get_master_cmd()
elif type(mpath) is bool:
    print("Setting {} as slave".format(identifier))
    my_path = slaves_path + "/" + identifier
    zk.ensure_path(my_path)
    s = "{} {}".format(get_ip_address(interface), port)
    zk.set(my_path, bytes(s, encoding="UTF-8"))
    value, stat = zk.get(master_path)
    cmd = get_slave_cmd(value.decode("UTF-8"))
else:
    raise RuntimeError("What the hell")

zk.stop()

#subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE).communicate()
try:
    print("Running: " + str(cmd))
    subprocess.call(cmd)
except:
    print("Exiting...")
    zk.start()
    #zk.delete(slaves_path + "/" + identifier)
    zk.delete(my_path)
    print("Deleted ZK entry")
    zk.stop()
    print("Done.")
