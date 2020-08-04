package util

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

var roomWaitGroup sync.WaitGroup
var roomSendMsgNum int

//player info
type player struct {
	id string
}

func (this *player) GetId() string {
	return this.id
}

//发消息
func (this *player) Send(msg interface{}) {
	roomSendMsgNum++
	roomWaitGroup.Done()
	//logger.DEBUG("send player%v msg:%v", this.playerId, msg)
}


func TestEmitter_Emit(t *testing.T) {
	msgNum := int32(10000)
	EventWriteChanSize = int(msgNum)

	emitter := NewEmitter()
	unionName := "unionRoom"
	unionRoom := emitter.On(unionName, true)
	unionRoom.Emit("0")
	PN := 10000
	for i := 1; i <= PN; i++ {
		P := &player{id: strconv.FormatInt(int64(i), 10)}
		emitter.On(unionName).Join(P)
	}

	for i := int32(1); i <= msgNum; i++ {
		roomWaitGroup.Add(PN)
		emitter.On(unionName).Emit(i)
	}
	roomWaitGroup.Wait()
	//unionRoom.close()
	fmt.Printf("累计发送信息：%v\n", roomSendMsgNum)
}

var workWaitGroup sync.WaitGroup

func myWorkerHandle(args interface{}) {
	//fmt.Println(args)
	workWaitGroup.Done()
}

func TestWorker_Emit(t *testing.T) {
	msgNum := int32(1000)

	work := NewWorker("chat", 4, myWorkerHandle)

	for i := int32(1); i <= msgNum; i++ {
		workWaitGroup.Add(1)
		work.Emit(i)
	}
	workWaitGroup.Wait()
	Stop(true)
}