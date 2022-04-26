package cosnet

import (
	"context"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/hwcer/cosgo/library/logger"
	"github.com/hwcer/cosgo/storage/cache"
	"net"
)

type NetIO interface {
	Read(head []byte) (Message, error)
	Write(msg Message) error
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

//Socket 基础网络连接
type Socket struct {
	io        NetIO
	stop      chan struct{} //stop
	agents    *Agents
	cwrite    chan Message //写入通道
	netType   NetType      //网络连接类型
	stopping  int8         //正在关闭
	heartbeat uint16       //heartbeat >=timeout 时被标记为超时
	cache.Data
}

//start 启动工作进程，status启动前状态
func (this *Socket) start() error {
	this.stop = make(chan struct{})
	this.agents.scc.CGO(this.readMsg)
	this.agents.scc.CGO(this.writeMsg)
	return nil
}

//close 内部关闭方式
func (this *Socket) close() {
	select {
	case <-this.stop:
	default:
		close(this.stop)
	}
}

//stopped 读写协程是否关闭或者正在关闭
func (this *Socket) stopped() bool {
	select {
	case <-this.stop:
		return true
	default:
		return false
	}
}

//Close 强制关闭,无法重连
func (this *Socket) Close(msg ...Message) {
	if len(msg) > 0 && !this.stopped() {
		this.stopping += 1
		for _, m := range msg {
			this.Write(m)
		}
	} else {
		this.close()
	}
}

func (this *Socket) Agents() *Agents {
	return this.agents
}

func (this *Socket) NetType() NetType {
	return this.netType
}
func (this *Socket) HasType(netType NetType) bool {
	return this.netType.Has(netType)
}
func (this *Socket) AnyType(netType NetType) bool {
	return this.netType.Any(netType)
}

//Heartbeat 每一次Heartbeat() heartbeat计数加1
func (this *Socket) Heartbeat() {
	this.heartbeat += 1
	if this.stopped() {
		this.agents.remove(this) //销毁
	} else if this.heartbeat >= Options.SocketConnectTime || (this.stopping > 0 && len(this.cwrite) == 0) {
		this.close()
	}
}

//KeepAlive 任何行为都清空heartbeat
func (this *Socket) KeepAlive() {
	this.heartbeat = 0
}

func (this *Socket) LocalAddr() net.Addr {
	return this.io.LocalAddr()
}
func (this *Socket) RemoteAddr() net.Addr {
	return this.io.RemoteAddr()
}

//Write 外部写入消息
func (this *Socket) Write(m Message) (re bool) {
	if m == nil {
		return false
	}
	if this.stopped() {
		//logger.Debug("SOCKET已经关闭无法写消息:%v", this.IId())
		return
	}
	select {
	case this.cwrite <- m:
		re = true
	default:
		logger.Debug(" 通道已满无法写消息:%v", this.Id())
		this.close()
	}
	return
}

//Json 发送Json数据
func (this *Socket) Json(code interface{}, index uint16, msg interface{}) (re bool) {
	var err error
	var data []byte
	if msg != nil {
		if data, err = json.Marshal(msg); err != nil {
			logger.Debug("socket Json error:%v", err)
			return false
		}
	}
	m := this.agents.Handler.New()
	m.Reset(data, code, index)
	return this.Write(m)
}

//Protobuf 发送Protobuf
func (this *Socket) Protobuf(code interface{}, index uint16, msg proto.Message) (re bool) {
	var err error
	var data []byte
	if msg != nil {
		if data, err = proto.Marshal(msg); err != nil {
			logger.Debug("socket Protobuf error; code:%v,err:%v", code, err)
			return false
		}
	}
	m := this.agents.Handler.New()
	m.Reset(data, code, index)
	return this.Write(m)
}

func (this *Socket) processMsg(socket *Socket, msg Message) {
	this.KeepAlive()
	//if msg.Head != nil && msg.Head.Flags.Has(message.FlagCompress) && msg.Data != nil {
	//	data, err := utils.GZipUnCompress(msg.Data)
	//	if err != nil {
	//		this.close()
	//		logger.Error("uncompress failed socket:%v err:%v", socket.IId(), err)
	//		return
	//	}
	//	msg.Data = data
	//	msg.Head.Flags.Leave(message.FlagCompress)
	//	msg.Head.size = uint32(len(msg.Data))
	//}
	logger.Debug("processMsg:%+v", msg)
	this.agents.Handler.Call(socket, msg)
}

func (this *Socket) readMsg(ctx context.Context) {
	defer this.close()
	head := make([]byte, this.agents.Handler.Size())
	for !this.stopped() {
		msg, err := this.io.Read(head)
		if err != nil {
			return
		}
		this.processMsg(this, msg)
	}
}

func (this *Socket) writeMsg(ctx context.Context) {
	defer this.close()
	defer this.io.Close()
	var msg Message
	for !this.stopped() {
		select {
		case <-this.stop:
			return
		case <-ctx.Done():
			return
		case msg = <-this.cwrite:
			if !this.writeMsgTrue(msg) {
				return
			}
		}
	}
}

func (this *Socket) writeMsgTrue(msg Message) bool {
	if this.io == nil {
		return false
	}
	if msg == nil {
		return true
	}
	if err := this.io.Write(msg); err != nil {
		logger.Debug("socket write error,IId:%v err:%v", this.Id(), err)
		return false
	}
	this.KeepAlive()
	return true
}
