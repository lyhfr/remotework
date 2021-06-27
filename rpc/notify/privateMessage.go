package notify

import (
	"net/rpc"

	"github.com/net-agent/remotework/rpc/notifyclient"
)

type PrivateMessageArgs struct {
	NameToken string
	Message   string
	Recivers  []string
}

type PrivateMessageReply struct {
}

func (n *Notify) PrivateMessage(args *PrivateMessageArgs, replay *PrivateMessageReply) error {
	sender, err := n.ParseNameToken(args.NameToken)
	if err != nil {
		return err
	}

	for _, reciver := range args.Recivers {
		go func(domain string) {
			it, loaded := n.clients.Load(domain)
			if !loaded {
				return
			}
			client, ok := it.(*rpc.Client)
			if !ok {
				return
			}

			client.Call("NotifyClient.PushNotify", &notifyclient.PushNotifyArgs{
				Sender:  sender,
				Message: args.Message,
			}, nil)
		}(reciver)
	}
	return nil
}
