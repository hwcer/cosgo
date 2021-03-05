package cosweb

type Options struct {
	SessionKey     string
	SessionType    []RequestDataType //存放SESSION KEY的方式
	SessionStorage storage           //Session数据存储器
}

func NewOptions() *Options {
	return &Options{
		SessionKey:  "CosWebSessId",
		SessionType: []RequestDataType{RequestDataTypeCookie, RequestDataTypeQuery},
	}
}
