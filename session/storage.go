// Package session 提供会话管理功能，支持内存和Redis存储
package session

// Storage 存储接口，定义了会话存储的核心方法
// 注意：
// 1. New: 同Create，创建新会话
// 2. Get: 验证TOKEN信息，获取会话数据
// 3. Create: 用户登录创建新session
// 4. Update: 更新session数据
// 5. Delete: 退出登录删除SESSION，关闭服务器时断开连接等
// 6. 可以通过实现此接口来扩展自定义存储后端
type Storage interface {
	New(data *Data) error                                             //同Create
	Get(id string) (data *Data, err error)                            //验证TOKEN信息
	Create(uuid string, value map[string]any) (data *Data, err error) //用户登录创建新session
	Update(data *Data, value map[string]any) error                    //更新session数据
	Delete(data *Data) error                                          //退出登录删除SESSION 	//关闭服务器时断开连接等
}

var listeners []func(data *Data)

// OnRelease 监听被释放的数据
func OnRelease(l func(data *Data)) {
	listeners = append(listeners, l)
}

func emitRelease(v *Data) {
	for _, l := range listeners {
		l(v)
	}
}
