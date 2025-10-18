package session

type Storage interface {
	New(data *Data) error                                             //同Create
	Get(id string) (data *Data, err error)                            //验证TOKEN信息
	Create(uuid string, value map[string]any) (data *Data, err error) //用户登录创建新session
	Update(data *Data, value map[string]any) error                    //更新session数据
	Delete(data *Data) error                                          //退出登录删除SESSION 	//关闭服务器时断开连接等
}
