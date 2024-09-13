package await

import "time"

var Default = New()

func init() {
	Default.Start(100, time.Second*5)
}

// Try 如果通道已满，立即放弃执行
func Try(handle Handle, args any) (any, error) {
	return Default.Try(handle, args)
}

// Call 同步调用handle并返回结果
func Call(handle Handle, args any) (any, error) {
	return Default.Call(handle, args)
}

// Sync 异步执行，不关心执行结果
// 也可以使用 Message.Done等待返回结果
func Sync(handle Handle, args any) *Message {
	return Default.Sync(handle, args)
}
