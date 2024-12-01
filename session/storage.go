package session

type Storage interface {
	Verify(token string) (data *Data, err error)                                 //验证TOKEN信息
	Create(uuid string, value map[string]any, ttl int64) (data *Data, err error) //用户登录创建新session
	Update(data *Data, value map[string]any, ttl int64) error                    //更新session数据
	Delete(data *Data) error                                                     //退出登录删除SESSION 	//关闭服务器时断开连接等
}
