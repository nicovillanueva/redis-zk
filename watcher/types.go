package main

import "gopkg.in/redis.v4"

type NodeReference struct {
	Host string
	Port string
}

func (nr NodeReference) GetAddress() string {
	return nr.Host + " " + nr.Port
}

func (nr NodeReference) GetDottedAddr() string {
	return nr.Host + ":" + nr.Port
}

type WatchOptions struct {
	Topic string
	// In the caught message, where will the host/port be?
	HostIndex int
	PortIndex int
	RedisOpts redis.Options
}
