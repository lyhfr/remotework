package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/net-agent/flex"
	"github.com/net-agent/remotework/rpc/notify"
)

func serveNotify(sw *flex.Switcher) {
	host, err := RegistHost(sw, "notify")
	if err != nil {
		log.Printf("regist host[notify] failed: %v\n", err)
		return
	}

	lsn, err := host.Listen(16)
	if err != nil {
		log.Printf("listen port[16] failed: %v\n", err)
		return
	}

	rpcsvr := rpc.NewServer()
	rpcsvr.Register(notify.New(host))
	rpcsvr.Accept(lsn)
}

func RegistHost(sw *flex.Switcher, domain string) (*flex.Host, error) {
	c1, c2 := net.Pipe()

	pc1 := flex.NewTcpPacketConn(c1)
	pc2 := flex.NewTcpPacketConn(c2)
	go sw.ServePacketConn(pc2)

	host, _, err := flex.UpgradeToHost(pc1, &flex.HostRequest{
		Domain: domain,
		Ctxid:  0,
		Mac:    "xx",
	}, true)

	return host, err
}
