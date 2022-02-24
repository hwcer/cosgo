package cosnet

type HandlerFunc func(*Socket, Message)

type Handler interface {
	New() Message                               //创建消息
	Size() int32                                //Head头长
	Call(*Socket, Message)                      //消息执行器
	Parse(head []byte) (Message, error)         //通过二进制头生成消息
	Handle(f HandlerFunc)                       //默认消息处理器
	Register(code interface{}, fun HandlerFunc) //注册消息处理函数
}

func NewHandle() Handler {
	handle := &HandlerDefault{
		dict:     make(map[interface{}]HandlerFunc),
		HeadSize: HeaderSize,
	}
	return handle
}

type HandlerDefault struct {
	dict     map[interface{}]HandlerFunc
	handle   HandlerFunc //默认消息处理器
	NewMsg   func() Message
	HeadSize int32
}

//New 创建MESSAGE
func (this *HandlerDefault) New() Message {
	if this.NewMsg != nil {
		return this.NewMsg()
	}
	msg := &message{}
	return msg
}

func (this *HandlerDefault) Size() int32 {
	return HeaderSize
}

func (this *HandlerDefault) Call(sock *Socket, msg Message) {
	code := msg.Code()
	if fn, ok := this.dict[code]; ok {
		fn(sock, msg)
	} else if this.handle != nil {
		this.handle(sock, msg)
	}
}
func (this *HandlerDefault) Parse(head []byte) (Message, error) {
	msg := this.New()
	if err := msg.Parse(head); err != nil {
		return nil, err
	}
	return msg, nil
}

func (this *HandlerDefault) Handle(f HandlerFunc) {
	this.handle = f
}

func (this *HandlerDefault) Register(code interface{}, fun HandlerFunc) {
	this.dict[code] = fun
}
