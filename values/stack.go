package values

import (
	"fmt"
	"runtime/debug"
)

type Stack struct {
	Message
	Stack string `json:"stack"`
}

func (this *Stack) String() string {
	s := this.Message.String()
	if this.Stack == "" {
		return s
	} else {
		return fmt.Sprintf("%v\n%v", s, this.Stack)
	}
}
func (this *Stack) Error() string {
	s := this.Message.Error()
	if this.Stack == "" {
		return s
	} else {
		return fmt.Sprintf("%v\n%v", s, this.Stack)
	}
}

// NewStack 获取一个带有堆栈的错误信息
func NewStack(code int, err interface{}, args ...interface{}) (r *Stack) {
	r = &Stack{}
	r.Message.Errorf(code, err, args...)
	r.Stack = string(debug.Stack())
	return
}
