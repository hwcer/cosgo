package cosnet

type MsgHandlerFunc func(Socket, *Message) bool

type MsgHandler interface {
	Register(int, MsgHandlerFunc)                   //注册消息
	OnConnect(Socket) bool                          //新的消息队列
	OnDisconnect(Socket)                            //消息队列关闭
	OnProcessMsg(Socket, *Message) bool             //默认的消息处理函数
	GetHandlerFunc(Socket, *Message) MsgHandlerFunc //根据消息获得处理函数
}

type DefMsgHandler struct {
	msgMap map[int]MsgHandlerFunc
}

func (r *DefMsgHandler) OnConnect(socket Socket) bool                  { return true }
func (r *DefMsgHandler) OnDisconnect(socket Socket)                    {}
func (r *DefMsgHandler) OnProcessMsg(socket Socket, msg *Message) bool { return true }
func (r *DefMsgHandler) GetHandlerFunc(socket Socket, msg *Message) MsgHandlerFunc {
	if f, ok := r.msgMap[int(msg.Head.Proto)]; ok {
		return f
	}
	return nil
}

func (r *DefMsgHandler) Register(act int, fun MsgHandlerFunc) {
	if r.msgMap == nil {
		r.msgMap = map[int]MsgHandlerFunc{}
	}
	r.msgMap[act] = fun
}
