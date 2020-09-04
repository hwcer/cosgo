package network

type HandlerFunc func(msgque Socket, msg *Message) bool

type IMsgHandler interface {
	OnNewMsgQue(msgque Socket) bool                         //新的消息队列
	OnDelMsgQue(msgque Socket)                              //消息队列关闭
	OnProcessMsg(msgque Socket, msg *Message) bool          //默认的消息处理函数
	GetHandlerFunc(msgque Socket, msg *Message) HandlerFunc //根据消息获得处理函数
}

type DefMsgHandler struct {
	msgMap  map[int]HandlerFunc
}

func (r *DefMsgHandler) OnNewMsgQue(socket Socket) bool                { return true }
func (r *DefMsgHandler) OnDelMsgQue(socket Socket)                     {}
func (r *DefMsgHandler) OnProcessMsg(socket Socket, msg *Message) bool { return true }
func (r *DefMsgHandler) GetHandlerFunc(socket Socket, msg *Message) HandlerFunc {
	if f, ok := r.msgMap[int(msg.Head.Act)]; ok {
		return f
	}
	return nil
}

func (r *DefMsgHandler) Register(act int, fun HandlerFunc) {
	if r.msgMap == nil {
		r.msgMap = map[int]HandlerFunc{}
	}
	r.msgMap[act] = fun
}