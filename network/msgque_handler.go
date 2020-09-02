package network

type HandlerFunc func(msgque IMsgQue, msg *Message) bool

type IMsgHandler interface {
	OnNewMsgQue(msgque IMsgQue) bool                         //新的消息队列
	OnDelMsgQue(msgque IMsgQue)                              //消息队列关闭
	OnProcessMsg(msgque IMsgQue, msg *Message) bool          //默认的消息处理函数
	GetHandlerFunc(msgque IMsgQue, msg *Message) HandlerFunc //根据消息获得处理函数
}

type DefMsgHandler struct {
	msgMap  map[int]HandlerFunc
}

func (r *DefMsgHandler) OnNewMsgQue(msgque IMsgQue) bool                { return true }
func (r *DefMsgHandler) OnDelMsgQue(msgque IMsgQue)                     {}
func (r *DefMsgHandler) OnProcessMsg(msgque IMsgQue, msg *Message) bool { return true }
func (r *DefMsgHandler) GetHandlerFunc(msgque IMsgQue, msg *Message) HandlerFunc {
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