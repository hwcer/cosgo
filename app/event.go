package app


type listener []func()

var emitter map[string]listener

type EventType int

const (
	EventType_Init EventType = iota    //初始化时
	EventType_Start     //开始启动(前)
	EventType_Runing    //运行时,启动后
	EventType_Stoped    //停止后
)


func On(name EventType,fun func())  {
	if emitter == nil{
		emitter = make(map[string]listener)
	}


}