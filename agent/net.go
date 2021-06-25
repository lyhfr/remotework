package main

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
)

func dial(addr string) (net.Conn, error) {
	network, hostname, port, err := parseAddr(addr)
	if err != nil {
		return nil, err
	}

	rawAddr := fmt.Sprintf("%v:%v", hostname, port)
	if network == "tcp" || network == "tcp4" || network == "tcp6" {
		return net.Dial(network, rawAddr)
	}

	if err != nil {
		return nil, err
	}
	return getHost().Dial(rawAddr)
}

func listen(addr string) (net.Listener, error) {
	network, hostname, port, err := parseAddr(addr)
	if err != nil {
		return nil, err
	}
	rawAddr := fmt.Sprintf("%v:%v", hostname, port)
	if network == "tcp" || network == "tcp4" || network == "tcp6" {
		return net.Listen(network, rawAddr)
	}

	return getHost().Listen(port)
}

func parseAddr(addr string) (network, host string, port uint16, err error) {
	re := regexp.MustCompile(`[a-zA-Z]+://.+`)
	if !re.MatchString(addr) {
		addr = "tcp://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		return "", "", 0, err
	}

	network = u.Scheme
	if network == "" {
		network = "tcp"
	}

	host, strPort, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", "", 0, err
	}
	if host == "" {
		host = "localhost"
	}

	intPort, err := strconv.ParseInt(strPort, 10, 16)
	port = uint16(intPort)

	return
}
