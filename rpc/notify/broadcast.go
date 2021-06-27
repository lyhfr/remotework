package notify

import (
	"net/rpc"

	"github.com/net-agent/remotework/rpc/notifyclient"
)

type BroadcastArgs struct {
	NameToken string
	Message   string
}

type BroadcastReply struct{}

// Broadcast 向所有在线连接广播消息
func (n *Notify) Broadcast(args *BroadcastArgs, reply *BroadcastReply) error {
	sender, err := n.ParseNameToken(args.NameToken)
	if err != nil {
		return err
	}

	n.clients.Range(func(key interface{}, val interface{}) bool {
		client, ok := val.(*rpc.Client)
		if ok {
			go client.Call("NotifyClient.PushNotify", &notifyclient.PushNotifyArgs{
				Sender:  sender,
				Message: args.Message,
			}, nil)
		}
		return true
	})

	return nil
}
