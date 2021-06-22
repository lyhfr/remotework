package main

import (
	"log"
	"net"
	"sync"

	"github.com/net-agent/flex"
)

func main() {
	var flags AgentFlags
	flags.Parse()

	// 读取配置
	log.Printf("> read config from '%v'\n", flags.ConfigFileName)
	config, err := NewConfig(flags.ConfigFileName)
	if err != nil {
		log.Fatal("load config failed: ", err)
	}

	// 创建连接
	log.Printf("> connect '%v'\n", config.Server.Address)
	conn, err := net.Dial("tcp4", config.Server.Address)
	if err != nil {
		log.Fatal("dial failed: ", err)
	}

	// 协议转换
	log.Printf("> upgrade to host, domain='%v'\n", config.Server.Vhost)
	host, err := flex.UpgradeToHost(conn, config.Server.Password, &flex.HostRequest{
		Domain: config.Server.Vhost,
		Mac:    "test-mac-token",
	})
	if err != nil {
		log.Fatal("upgrade failed: ", err)
	}

	log.Printf("> host created, ip=%v\n", host.IP())

	var wg sync.WaitGroup

	// 开启服务
	log.Println("-------------------------------------------------------------------------")
	log.Println("state index        type                   listen                   target")
	log.Println("-------------------------------------------------------------------------")

	for i := 0; i < len(config.Services); i++ {
		svc := config.Services[i]

		enable := "stop"
		if svc.Enable {
			enable = "run"
		}

		log.Printf("%5v %5v %11v %24v %24v\n", enable, i, svc.Type, svc.Param["listen"], svc.Param["target"])
		wg.Add(1)
		go func(svc *ServiceInfo) {
			svc.Run(&wg, host)
		}(&svc)
	}

	log.Println("-------------------------------------------------------------------------")

	wg.Wait()
	log.Println("agent stopped")
}
