package cosnet

import (
	"fmt"
	"github.com/hwcer/cosgo/cosnet/message"
	"github.com/hwcer/cosgo/utils"
	"testing"
	"time"
)

var msg *message.Message
var sockets = NewSockets(&HandlerDefault{}, 10)

func init() {
	msg = &message.Message{Head: &message.Head{Index: 1}}
}

func TestSocket(t *testing.T) {
	//address := "0.0.0.0:3100"
	//for i := 1; i <= 1; i++ {
	//	NewTcpClient(sockets, address)
	//}
	//sockets.scc.CGO(startSocketHeartbeat)
	//sockets.Wait()
	log(0)
	for i := 1; i <= 22; i++ {
		sockets.New()
		log(i)
	}
}

func log(i int) {
	fmt.Printf("%v ,size:%v,index:%v,dirty:%v \n", i, sockets.Size(), sockets.dirty.index, sockets.dirty.list)
}

func startSocketHeartbeat(stop chan struct{}) {
	t := time.Second * 10
	ticker := time.NewTimer(t)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			utils.Try(heartbeat)
			ticker.Reset(t)
		}
	}
}

func heartbeat() {
	sockets.Broadcast(msg.NewMsg(123, []byte("321")), nil)
}
