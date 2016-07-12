from kazoo.client import KazooClient
import uuid, os, socket, fcntl, struct, shlex, subprocess

BASE_PATH="/redis"
MASTER_PATH="{}/master".format(BASE_PATH)
SENTINELS_PATH="{}/sentinels".format(BASE_PATH)
SENTINEL_CONF="/etc/redis/sentinel.conf"
INTERFACE=os.environ.get("NET_IFACE") or 'eth0'
IDENTIFIER=str(uuid.uuid4())

def get_ip_address(ifname):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    return socket.inet_ntoa(fcntl.ioctl(
        s.fileno(),
        0x8915,  # SIOCGIFADDR
        struct.pack('256s', bytes(ifname, 'utf-8'))
    )[20:24])

zk = KazooClient(hosts=os.environ.get("ZK_HOSTS") or '127.0.0.1:2181')
zk.start()
value, stat = zk.get(MASTER_PATH)
master = value.decode("UTF-8")
with open(SENTINEL_CONF, 'w+') as f:
    content = f.read()
    f.seek(0)
    f.write('sentinel monitor master1 {} 2\n'.format(master.rstrip()))
    f.write(content)
    print("Wrote master '{}' to conf file".format(master))

ipaddr = get_ip_address(INTERFACE)
hostport = "{} {}".format(ipaddr, os.environ.get("PORT0"))
zk.create(SENTINELS_PATH + "/" + IDENTIFIER, bytes(hostport, encoding='UTF-8'), makepath=True)
zk.stop()

cmd = "redis-sentinel {} --bind {} --port {}".format(SENTINEL_CONF, ipaddr, os.environ.get("PORT0"))
print("Running {}".format(cmd))
cmd = shlex.split(cmd)
subprocess.call(cmd)
