package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/net-agent/remotework/agent"
	"github.com/net-agent/socks"
)

type QuickVisit struct {
	hub       *agent.NetHub
	listenURL string // example: "tcp://localhost:1000", "flex://0:1001"
	targetURL string // example: "tcp://agentname:secret@localhost:3389"
	logName   string

	listener   net.Listener
	dialer     agent.QuickDialer
	upgrader   *socks.ProxyInfo
	targetAddr string
	mut        sync.Mutex

	actives int32
	dones   int32
}

func NewQuickVisit(hub *agent.NetHub, listenURL, targetURL, logName string) *QuickVisit {
	return &QuickVisit{
		hub:       hub,
		listenURL: listenURL,
		targetURL: targetURL,
		logName:   logName,
	}
}
func (s *QuickVisit) Report() agent.ReportInfo {
	return agent.ReportInfo{
		Name:    s.Name(),
		State:   "uninit",
		Listen:  s.listenURL,
		Target:  s.targetURL,
		Actives: s.actives,
		Dones:   s.dones,
	}
}

func (s *QuickVisit) Name() string {
	if s.logName != "" {
		return s.logName
	}
	return "portp"
}
func (s *QuickVisit) Network() string { return "tcp4" }

func (ctx *QuickVisit) Init() error {
	// init network domain dialer
	u, err := url.Parse(ctx.targetURL)
	if err != nil {
		return err
	}
	network := u.Scheme
	domain := u.User.Username()
	vals := url.Values{}
	vals.Add("secret", QuickSecret)
	dial, _ := ctx.hub.URLDialer(fmt.Sprintf("%v://%v:%v?%v", network, domain, QuickPort, vals.Encode()))
	ctx.dialer = dial
	ctx.targetAddr = u.Host

	// init socks5 proxy upgrader
	secret, ok := u.User.Password()
	if !ok {
		return errors.New("parse secret failed")
	}
	ctx.upgrader = &socks.ProxyInfo{
		Network:  "tcp4",
		Address:  "", // 只用到upgrader，不需要创建连接
		NeedAuth: true,
		Username: "", // 由dialer进行进行校验
		Password: secret,
	}

	// init listener
	if err = ctx.Update(); err != nil {
		return err
	}

	return nil
}

func (ctx *QuickVisit) Update() error {
	ctx.mut.Lock()
	defer ctx.mut.Unlock()

	l, err := ctx.hub.ListenURL(ctx.listenURL)
	if err != nil {
		return err
	}
	if ctx.listener != nil {
		ctx.listener.Close()
	}
	ctx.listener = l
	return nil
}
func (ctx *QuickVisit) getlistener() net.Listener {
	ctx.mut.Lock()
	defer ctx.mut.Unlock()
	return ctx.listener
}

func (ctx *QuickVisit) Start() error {
	if ctx.listener == nil || ctx.upgrader == nil || ctx.dialer == nil {
		return errors.New("init failed")
	}

	l := ctx.getlistener()
	for {
		c1, err := l.Accept()
		if err != nil {
			if l != ctx.getlistener() {
				l = ctx.getlistener()
				if l != nil {
					log.Printf("[%v] listener updated\n", ctx.logName)
					continue
				}
			}
			return err
		}

		go ctx.serve(c1)
	}
}

func (ctx *QuickVisit) serve(c1 net.Conn) {
	atomic.AddInt32(&ctx.actives, 1)
	defer func() {
		c1.Close()
		atomic.AddInt32(&ctx.actives, -1)
		atomic.AddInt32(&ctx.dones, 1)
	}()

	// connect to network/domain
	c2, err := ctx.dialer()
	if err != nil {
		return
	}
	defer c2.Close()

	// upgrade socks5 request
	c2, err = ctx.upgrader.Upgrade(c2, ctx.targetAddr)
	if err != nil {
		return
	}

	link(c1, c2)
}

func link(c1, c2 io.ReadWriteCloser) (c1ReadN, c1WriteN int64, err error) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		c1WriteN, _ = io.Copy(c1, c2)
		c1.Close()
		c2.Close()
		wg.Done()
	}()

	c1ReadN, _ = io.Copy(c2, c1)
	c1.Close()
	c2.Close()

	wg.Wait()

	return
}

func (ctx *QuickVisit) Close() error {
	if ctx.listener != nil {
		ctx.listener.Close()
	}
	return nil
}
