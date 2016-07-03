from kazoo.client import KazooClient
import os, shlex, subprocess, socket, fcntl, struct

BASE_PATH="/redis"
MASTER_PATH="{}/master".format(BASE_PATH)
SLAVES_PATH="{}/slaves".format(BASE_PATH)
PORTS_PATH="{}/portlocks".format(BASE_PATH)
INTERFACE=os.environ.get("NET_IFACE") or 'eth0'
IDENTIFIER=str(os.getpid())

zk = KazooClient(hosts=os.environ.get("ZK_HOSTS") or '127.0.0.1:2181')

def get_master_cmd():
    c = "redis-server --port {}".format(os.environ.get("PORT0") or "6379")
    return shlex.split(c)

def get_slave_cmd(master):
    c = "redis-server --port {} --slaveof {}".format(os.environ.get("PORT0") or "6379", master)
    return shlex.split(c)


def get_ip_address(ifname):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    return socket.inet_ntoa(fcntl.ioctl(
        s.fileno(),
        0x8915,  # SIOCGIFADDR
        struct.pack('256s', bytes(ifname, 'utf-8'))
    )[20:24])



zk.start()

mpath = zk.ensure_path(MASTER_PATH)
print("MPath is: {}".format(mpath))
if type(mpath) is str:
    with zk.Lock(MASTER_PATH, "{}-master".format(IDENTIFIER)) as lock:
        print("setting {} as master".format(IDENTIFIER))
        s = "{} {}".format(get_ip_address(INTERFACE), os.environ.get("PORT0"))
        zk.set(MASTER_PATH, bytes(s, encoding="UTF-8"))
        cmd = get_master_cmd()
elif type(mpath) is bool:
    print("setting {} as slave".format(IDENTIFIER))
    zk.create(SLAVES_PATH + "/" + str(IDENTIFIER), makepath=True)
    value, stat = zk.get(MASTER_PATH)
    cmd = get_slave_cmd(value.decode("UTF-8"))
else:
    raise RuntimeError("What the hell")

zk.stop()

print("Running: " + str(cmd))
subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE).communicate()