package notify

import (
	"sync"

	"github.com/net-agent/flex"
)

type Notify struct {
	host    *flex.Host
	clients sync.Map // map[name string]*rpc.Client
}

func New(host *flex.Host) *Notify {
	return &Notify{
		host: host,
	}
}

func (n *Notify) ParseNameToken(token string) (string, error) {
	return token, nil
}
