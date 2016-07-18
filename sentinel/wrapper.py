from kazoo.client import KazooClient
import uuid, os, socket, fcntl, struct, shlex, subprocess

zk_hosts = os.environ.get("ZK_HOSTS") or '127.0.0.1:2181'
base_path = "/redis"
master_path = "{}/master".format(base_path)
sentinels_path = "{}/sentinels".format(base_path)
sentinel_conf = "/etc/redis/sentinel.conf"

interface = os.environ.get("NET_IFACE") or 'eth0'
port = os.environ.get("PORT0") or '26379'
identifier = str(uuid.uuid4())


def get_ip_address(ifname):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    return socket.inet_ntoa(fcntl.ioctl(
        s.fileno(),
        0x8915,  # SIOCGIFADDR
        struct.pack('256s', bytes(ifname, 'utf-8'))
    )[20:24])

zk = KazooClient(hosts=zk_hosts)
zk.start()
value, stat = zk.get(master_path)
master = value.decode("UTF-8")
with open(sentinel_conf, 'w+') as f:
    content = f.read()
    f.seek(0)
    f.write('sentinel monitor master1 {} 2\n'.format(master.rstrip()))
    f.write(content)
    print("Wrote master '{}' to conf file".format(master))

ipaddr = get_ip_address(interface)
hostport = "{} {}".format(ipaddr, port)
my_path = sentinels_path + "/" + identifier
zk.create(my_path, bytes(hostport, encoding='UTF-8'), makepath=True)
zk.stop()


try:
    cmd = "redis-sentinel {} --bind {} --port {}".format(sentinel_conf, ipaddr, port)
    print("Running {}".format(cmd))
    cmd = shlex.split(cmd)
    subprocess.call(cmd)
except:
    print("Exiting...")
    zk.start()
    #zk.delete(slaves_path + "/" + identifier)
    zk.delete(my_path)
    print("Deleted ZK entry")
    zk.stop()
    print("Done.")
