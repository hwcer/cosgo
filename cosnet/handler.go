package cosnet

type HandlerMessageFunc func(Socket, *Message) bool

type Handler interface {
	Message(Socket, *Message) bool //消息处理函数
}

func NewHandlerDefault() *HandlerDefault {
	return &HandlerDefault{
		Handle: make(map[int]HandlerMessageFunc),
	}
}

type HandlerDefault struct {
	Handle map[int]HandlerMessageFunc
}

func (r *HandlerDefault) Message(sock Socket, msg *Message) bool {
	if f, ok := r.Handle[int(msg.Head.Proto)]; ok {
		return f(sock, msg)
	}
	return true
}

func (r *HandlerDefault) Register(act int, fun HandlerMessageFunc) {
	if r.Handle == nil {
		r.Handle = map[int]HandlerMessageFunc{}
	}
	r.Handle[act] = fun
}
