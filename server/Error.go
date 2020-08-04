package server

const (
	ErrMsg_NAME_SUCCESS = "SUCCESS"
	ErrMsg_NAME_DEFCODE = "DefErrCode"
)

//字符串转错误码
var errMsg map[string]int

func GetErrCode(err string) int {
	code, ok := errMsg[err]
	if ok {
		return code
	} else {
		return errMsg[ErrMsg_NAME_DEFCODE]
	}
}

func SetErrCode(err string,code int)  {
	errMsg[err] = code
}

func init()  {
	errMsg = make(map[string]int)
	SetErrCode(ErrMsg_NAME_SUCCESS,0)
	SetErrCode(ErrMsg_NAME_DEFCODE,-999999)
}