package agent

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/net-agent/flex/v2/node"
	"github.com/net-agent/flex/v2/packet"
	"github.com/net-agent/flex/v2/switcher"
	"github.com/net-agent/remotework/utils"
)

type Config struct {
	Agents    []AgentInfo     `json:"agents"`
	Portproxy []PortproxyInfo `json:"portproxy"`
	Socks5    []Socks5Info    `json:"socks5"`
	RDP       []RDPInfo       `json:"rdp"`
}

func NewConfig(configFileName string) (*Config, error) {
	cfg := &Config{}
	var err error
	switch strings.ToLower(path.Ext(configFileName)) {
	case ".json":
		err = utils.LoadJSONFile(configFileName, cfg)
	case ".toml":
		err = utils.LoadTomlFile(configFileName, cfg)
	default:
		err = fmt.Errorf("config file [%s] not support, must be json or toml", configFileName)
	}
	return cfg, err
}

type ServerInfo struct {
	Listen   string `json:"listen"`   // 监听的地址
	Password string `json:"password"` // 校验连接的密码
	WsEnable bool   `json:"wsEnable"` // 是否启用Websocket
	WsPath   string `json:"wsPath"`   // Websocket路径
}

type AgentInfo struct {
	Enable     bool   `json:"enable"`
	Network    string `json:"network"`  // 网络名称，不能为tcp、tcp4、tcp6
	Address    string `json:"address"`  // 服务端地址
	Password   string `json:"password"` // 连接服务的密码
	Domain     string `json:"domain"`   // 独立域名（不能重复）
	WsEnable   bool   `json:"wsEnable"` // 是否为Websocket服务
	Wss        bool   `json:"wss"`      // 是否为wss协议
	WsPath     string `json:"wsPath"`   // Websocket路径
	QuickTrust Trust  `json:"trust"`
}

type Trust struct {
	Enable    bool              `json:"enable"`
	WhiteList map[string]string `json:"whiteList"`
}

func (agent *AgentInfo) GetConnectFn() ConnectFunc {
	macs, _ := getMacAddr()
	macStr := strings.Join(macs, " ")

	if agent.WsEnable {
		return agent.getWsConnectFn(macStr)
	}

	return agent.getTcpConnectFn(macStr)
}

func (agent *AgentInfo) getWsConnectFn(mac string) ConnectFunc {

	u := url.URL{
		Scheme: "ws",
		Host:   agent.Address,
		Path:   agent.WsPath,
	}
	if agent.Wss {
		u.Scheme = "wss"
	}
	wsurl := u.String()

	return func() (*node.Node, error) {
		log.Printf("connect to '%v'\n", wsurl)
		c, _, err := websocket.DefaultDialer.Dial(wsurl, nil)
		if err != nil {
			return nil, err
		}
		pc := packet.NewWithWs(c)
		node, err := switcher.UpgradeToNode(
			pc,
			agent.Domain,
			mac,
			agent.Password,
		)
		if err != nil {
			c.Close()
			return nil, err
		}
		return node, nil
	}
}

func (agent *AgentInfo) getTcpConnectFn(mac string) ConnectFunc {

	return func() (*node.Node, error) {
		log.Printf("connect to '%v'\n", agent.Address)
		c, err := net.Dial("tcp4", agent.Address)
		if err != nil {
			return nil, err
		}
		pc := packet.NewWithConn(c)
		node, err := switcher.UpgradeToNode(
			pc,
			agent.Domain,
			mac,
			agent.Password,
		)
		if err != nil {
			c.Close()
			return nil, err
		}
		return node, nil
	}
}

type PortproxyInfo struct {
	ListenURL string `json:"listen"`
	TargetURL string `json:"target"`
	LogName   string `json:"log"`
}

type Socks5Info struct {
	ListenURL string `json:"listen"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	LogName   string `json:"log"`
}

type RDPInfo struct {
	ListenURL string `json:"listen"`
	LogName   string `json:"log"`
}

func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}
