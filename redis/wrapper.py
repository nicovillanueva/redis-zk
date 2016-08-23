from kazoo.client import KazooClient
import os, shlex, subprocess, socket, fcntl, struct, uuid

zk_hosts = os.environ.get("ZK_HOSTS") or '127.0.0.1:2181'
bind_ip = os.environ.get("BIND_IP")
interface = os.environ.get("NET_IFACE") or 'eth0'
port = os.environ.get("PORT0") or "6379"

# uid = str(uuid.uuid4())
is_master = False

zk = KazooClient(hosts=zk_hosts)


def setup_zk():
    assert zk.connected
    zk.ensure_path('/redis/master')
    #zk.ensure_path('/redis/members')
    #zk.ensure_path('/redis/roles/master')
    #zk.ensure_path('/redis/roles/slaves')


def register(role):
    assert zk.connected
    s = "{} {}".format(get_ip_address(interface), port)
    #print("Registering ID: %s".format(uid))
    #zk.create('/redis/members/' + uid)
    # zk.create('/redis/roles/' + role + '/' + uid)
    # zk.set('/redis/roles/' + role + '/' + uid, bytes(s, encoding='UTF-8'))
    if role == 'master':
        #zk.set('/redis/roles/master', bytes(s, encoding='UTF-8'))
        zk.set('/redis/master', bytes(s, encoding='UTF-8'))
    elif role == 'slaves':
        #zk.create('/redis/roles/slaves/' + uid)
        #zk.set('/redis/roles/slaves/' + uid, bytes(s, encoding='UTF-8'))
        print('Not registering slave')
    else:
        raise RuntimeException('Unrecognized role:' + role)



def get_master_host():
    assert zk.connected
    # master_id = zk.get_children('/redis/roles/master')[0]
    # master_host, _ = zk.get('/redis/roles/master/' + master_id)
    #master_host, _ = zk.get('/redis/roles/master')
    master_host, _ = zk.get('/redis/master')
    return master_host


def get_master_cmd():
    c = "redis-server /etc/redis/redis.conf --port {}".format(port)
    return shlex.split(c)


def get_slave_cmd(master):
    c = "redis-server /etc/redis/redis.conf --port {} --slaveof {}".format(port, master)
    return shlex.split(c)


def get_ip_address(ifname):
    if bind_ip is not None:
        return bind_ip
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    return socket.inet_ntoa(fcntl.ioctl(
        s.fileno(),
        0x8915,  # SIOCGIFADDR
        struct.pack('256s', bytes(ifname, 'utf-8'))
    )[20:24])


zk.start()
setup_zk()
# if len(zk.get_children('/redis/roles/master')) == 0:
# if zk.get('/redis/roles/master')[0] == b'':
if zk.get('/redis/master')[0] == b'':
    #with zk.Lock('/redis/roles/master', "{}-master".format(uid)) as lock:
    #with zk.Lock('/redis/master', "{}-master".format(uid)) as lock:
    with zk.Lock('/redis/master', "master-update") as lock:
        #print("Setting {} as master".format(uid))
        register('master')
        is_master = True
        cmd = get_master_cmd()
else:
    #print("Setting {} as slave".format(uid))
    register('slaves')
    mhost = get_master_host()
    cmd = get_slave_cmd(mhost.decode("UTF-8"))

zk.stop()

try:
    print("Running: " + str(cmd))
    subprocess.call(cmd)
except:
    print("Exiting...")
    #zk.start()
    #zk.delete('/redis/members/' + uid)
    #if is_master:
    #    zk.set('/redis/roles/master', b'')
    #else:
    #    zk.delete('/redis/roles/slaves/' + uid)
    #print("Deleted ZK entries")
    #zk.stop()
    #print("Done.")
