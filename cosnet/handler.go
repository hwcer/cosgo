package cosnet

type Handler interface {
	OnMessage(Socket, *Message) bool //消息处理函数
	OnConnect(Socket) bool           //新的消息队列
	OnDisconnect(Socket)             //消息队列关闭
}

type HandlerFunc func(Socket, *Message) bool

type HandlerDefault struct {
	handle map[int]HandlerFunc
}

func (r *HandlerDefault) OnMessage(socket Socket, msg *Message) bool {

	return true
}
func (r *HandlerDefault) OnConnect(socket Socket) bool {
	return true
}

func (r *HandlerDefault) OnDisconnect(socket Socket) {

}

func (r *HandlerDefault) GetHandlerFunc(socket Socket, msg *Message) HandlerFunc {
	if f, ok := r.handle[int(msg.Head.Proto)]; ok {
		return f
	}
	return nil
}

func (r *HandlerDefault) Register(act int, fun HandlerFunc) {
	if r.handle == nil {
		r.handle = map[int]HandlerFunc{}
	}
	r.handle[act] = fun
}
