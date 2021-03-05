package cosweb

type storage interface {
	Get(string) map[string]interface{}     //获取&加锁
	Set(string, string, interface{})       //设置
	Save(string)                           //保存&解锁
	Create(string, map[string]interface{}) //创建以name为名的表
	Delete(string)                         //删除以name为名的表
}

type Session struct {
	s *Server
	c *Context
}

func NewSession(s *Server, c *Context) *Session {
	return &Session{s: s, c: c}
}

func (this *Session) reset() {

}
func (this *Session) release() {

}

func (this *Session) Start(level ...int8) {

}
