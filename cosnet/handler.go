package cosnet

import "github.com/hwcer/cosgo/cosnet/message"

type HandlerMessageFunc func(Socket, *message.Message)

type Handler interface {
	Message(Socket, *message.Message) //消息处理函数
}

func NewHandlerDefault() *HandlerDefault {
	return &HandlerDefault{
		Handle: make(map[int]HandlerMessageFunc),
	}
}

type HandlerDefault struct {
	Handle map[int]HandlerMessageFunc
}

func (r *HandlerDefault) Message(sock Socket, msg *message.Message) {
	if f, ok := r.Handle[int(msg.Head.Code)]; ok {
		f(sock, msg)
	}
}

func (r *HandlerDefault) Register(act int, fun HandlerMessageFunc) {
	if r.Handle == nil {
		r.Handle = map[int]HandlerMessageFunc{}
	}
	r.Handle[act] = fun
}
