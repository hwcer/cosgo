package handler

const (
	ErrMsg_NAME_SUCCESS = "SUCCESS"
	ErrMsg_NAME_DEFCODE = "DefErrCode"
)

//字符串转错误码
var ErrMsg = map[string]int{
	"SUCCESS" : 0,
	"DefErrCode":-999999,
}

func GetErrCode(err string) int {
	code, ok := ErrMsg[err]
	if ok {
		return code
	} else {
		return ErrMsg[ErrMsg_NAME_DEFCODE]
	}
}

func SetErrCode(err string,code int)  {
	ErrMsg[err] = code
}