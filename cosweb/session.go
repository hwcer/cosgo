package cosweb

type storage interface {
	Get(string) interface{}                //获取
	Set(string, interface{})               //设置
	Start() error                          //启动检查数据
	Create(string, map[string]interface{}) //创建以name为名的表
	Lock() bool                            //乐观锁
	UnLock()                               //解锁
}

type Session struct {
	c       *Context
	storage storage
}

func NewSession(c *Context) *Session {
	return &Session{c: c}
}

func (this *Session) reset(c *Context) {
	this.c = c
}
func (this *Session) release() {
	this.c = nil
}
