package notify

import (
	"errors"
	"fmt"
	"net/rpc"
)

type JoinArgs struct {
	NameToken string
}
type JoinReply struct {
}

// Join 注册
// todo: 此处应该与Switcher共享SSO认证
func (n *Notify) Join(args *JoinArgs, reply *JoinReply) error {
	domain, err := n.ParseNameToken(args.NameToken)
	if err != nil {
		return err
	}

	stream, err := n.host.Dial(fmt.Sprintf("%v:15", domain))
	if err != nil {
		return err
	}

	client := rpc.NewClient(stream)
	_, loaded := n.clients.LoadOrStore(domain, client)
	if loaded {
		return errors.New("name exists")
	}
	return nil
}
