package main

//var Members type map[string]nodeReference

type nodeReference struct {
    Host string
    Port string
}

func (nr nodeReference) GetAddress() string {
    return nr.Host + " " + nr.Port
}

func (nr nodeReference) GetDottedAddr() string {
    return nr.Host + ":" + nr.Port
}

type WatchOptions struct {
    Topic string
    // In the caught message, where will the host/port be?
    HostIndex int
    PortIndex int
}
