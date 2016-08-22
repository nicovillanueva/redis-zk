from kazoo.client import KazooClient
import uuid, os, socket, fcntl, struct, shlex, subprocess, time, sys

zk_hosts = os.environ.get("ZK_HOSTS") or '127.0.0.1:2181'
sentinel_conf = "/etc/redis/sentinel.conf"
sentinel_quorum = os.environ.get("SENTINEL_QUORUM") or 2

interface = os.environ.get("NET_IFACE") or 'eth0'
port = os.environ.get("PORT0") or '26379'

sid = str(uuid.uuid4())


def get_ip_address(ifname):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    return socket.inet_ntoa(fcntl.ioctl(
        s.fileno(),
        0x8915,  # SIOCGIFADDR
        struct.pack('256s', bytes(ifname, 'utf-8'))
    )[20:24])


def get_master_location():
    assert zk.connected
    elapsed = 0
    # TODO: Better way to wait?
    # while zk.exists('/redis/roles/master') is None or zk.get('/redis/roles/master')[0] == b'':
    while zk.exists('/redis/master') is None or zk.get('/redis/master')[0] == b'':
        if elapsed > 30:
            print('No master found in 30 seconds. Aborting.')
            zk.stop()
            sys.exit(1)
        print('No master yet. Waiting.')
        time.sleep(1)
        elapsed += 1
    #master_id = zk.get_children('/redis/roles/master')[0]
    # master_host = zk.get('/redis/roles/master')[0]
    master_host = zk.get('/redis/master')[0]
    return master_host.decode('UTF-8')  # , master_id


def setup_zk():
    assert zk.connected
    #zk.ensure_path('/redis/roles/sentinels')
    #zk.ensure_path('/redis/members')
    zk.ensure_path('/redis/sentinels')


zk = KazooClient(hosts=zk_hosts)
zk.start()
#setup_zk()
m_loc = get_master_location()
with open(sentinel_conf, 'w+') as f:
    content = f.read()
    f.seek(0)
    f.write('sentinel monitor master1 {} {}\n'.format(m_loc.rstrip(), sentinel_quorum))
    f.write(content)
    print("Wrote master '{}' to conf file".format(m_loc))

ipaddr = get_ip_address(interface)
hostport = "{} {}".format(ipaddr, port)
#zk.create("/redis/roles/sentinels/" + sid, bytes(hostport, encoding='UTF-8'))
#zk.create("/redis/members/" + sid, bytes(hostport, encoding='UTF-8'))
zk.create("/redis/sentinels/" + sid, bytes(hostport, encoding='UTF-8'), makepath=True)
zk.stop()


try:
    cmd = "redis-sentinel {} --bind {} --port {}".format(sentinel_conf, ipaddr, port)
    print("Running {}".format(cmd))
    cmd = shlex.split(cmd)
    subprocess.call(cmd)
except:
    print("Exiting...")
    zk.start()
    #zk.delete("/redis/roles/sentinels/" + sid)
    #zk.delete("/redis/members/" + sid)
    zk.delete("/redis/sentinels/" + sid)
    #print("Deleted ZK entries")
    print("Deleted ZK entry")
    zk.stop()
    print("Done.")
