package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/net-agent/remotework/agent"
)

type Portproxy struct {
	mnet *agent.MixNet
	info agent.ServiceInfo

	closer       io.Closer
	listen       string
	target       string
	targetDialer agent.Dialer
}

func NewPortproxy(mnet *agent.MixNet, info agent.ServiceInfo) *Portproxy {
	target := info.Param["target"]
	dialer, err := mnet.URLDialer(target)
	if err != nil {
		panic(fmt.Sprintf("init portproxy failed, make dialer failed: %v", err))
	}
	return &Portproxy{
		mnet: mnet,
		info: info,

		listen:       info.Param["listen"],
		target:       target,
		targetDialer: dialer,
	}
}

func (p *Portproxy) Info() string {
	if p.info.Enable {
		return agent.Green(fmt.Sprintf("%11v %24v %24v", p.info.Type, p.listen, p.target))
	}
	return agent.Yellow(fmt.Sprintf("%11v %24v", p.info.Type, "disabled"))
}

func (p *Portproxy) Start(wg *sync.WaitGroup) error {
	if !p.info.Enable {
		return errors.New("service disabled")
	}

	l, err := p.mnet.ListenURL(p.listen)
	if err != nil {
		return err
	}

	p.closer = l

	runsvc(p.info.Name(), wg, func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go p.serve(conn)
		}
	})
	return nil
}

func (p *Portproxy) Close() error {
	return p.closer.Close()
}

func (p *Portproxy) serve(c1 net.Conn) {
	var dialer string
	if s, ok := c1.(interface{ Dialer() string }); ok {
		dialer = "flex://" + s.Dialer()
	} else {
		dialer = "tcp://" + c1.RemoteAddr().String()
	}

	c2, err := p.targetDialer()
	if err != nil {
		log.Printf("[%v] dial listen='%v' failed. %v\n", p.info.Type, p.listen, err)
		c1.Close()
		return
	}

	log.Printf("[%v] connect, dialer='%v' listen='%v'\n", p.info.Type, dialer, p.listen)

	link(c1, c2)
}
