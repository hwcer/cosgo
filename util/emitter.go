package util

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type eventMessageType uint8

const (
	eventMessageType_Emit eventMessageType = iota
	eventMessageType_Join
	eventMessageType_Leave
)

//房间实体
type Event struct {
	stop     int32
	name     string
	cwrite   chan *eventMessage
	handle   map[eventMessageType]func(message *eventMessage)
	listener map[string]eventListener
}

//Emitter 房间管理器，并发安全
type Emitter struct {
	mutex sync.Mutex
	rooms map[string]*Event
}

//内部消息
type eventMessage struct {
	Act  eventMessageType
	Data interface{}
}

//外部监听者
type eventListener interface {
	GetId() string    //监听者唯一ID，如playerId
	Send(interface{}) //消息接收
}

func newEventMessage(act eventMessageType, data interface{}) *eventMessage {
	return &eventMessage{
		Act:  act,
		Data: data,
	}
}

//创建一个独立的事件房间
func NewEvent(name string) *Event {
	return &Event{
		name:     name,
		cwrite:   make(chan *eventMessage, EventWriteChanSize),
		handle:   make(map[eventMessageType]func(message *eventMessage)),
		listener: make(map[string]eventListener),
	}
}

//创建事件管理器
func NewEmitter() *Emitter {
	return &Emitter{
		rooms: make(map[string]*Event),
	}
}

//Get 获取一个现有事件，如果不存在根据autoCreate返回nil或者创建新事件之后返回
func (this *Emitter) On(name string, autoCreate ...bool) *Event {
	if len(autoCreate) > 0 && autoCreate[0] {
		return this.Create(name)
	}
	this.mutex.Lock()
	room := this.rooms[name]
	this.mutex.Unlock()
	return room
}

//Create 创建事件,如果已经存在，直接返回已经存在的房间
func (this *Emitter) Create(name string) *Event {
	this.mutex.Lock()
	room := this.rooms[name]
	if room == nil {
		room = NewEvent(name)
		Go(room.start)
		this.rooms[name] = room
	}
	this.mutex.Unlock()
	return room
}

//Close 关闭一个房间
func (this *Emitter) Close(name string) {
	this.mutex.Lock()
	room := this.rooms[name]
	if room != nil {
		delete(this.rooms, name)
	}
	this.mutex.Unlock()
	if room != nil {
		room.Close()
	}
}

func (this *Event) start(ctx context.Context) {
	//注册事件
	this.handle[eventMessageType_Emit] = this.emit
	this.handle[eventMessageType_Join] = this.join
	this.handle[eventMessageType_Leave] = this.leave

	for this.stop == 0 && !IsStop() {
		select {
		case <-ctx.Done():
			this.Close()
		case msg := <-this.cwrite:
			if fn := this.handle[msg.Act]; fn != nil {
				fn(msg)
			}
		}
	}
	close(this.cwrite)
}

func (this *Event) emit(msg *eventMessage) {
	for _, listener := range this.listener {
		listener.Send(msg.Data)
	}
}
func (this *Event) join(msg *eventMessage) {
	listener := msg.Data.(eventListener)
	listenerId := listener.GetId()
	this.listener[listenerId] = listener
}

func (this *Event) leave(msg *eventMessage) {
	listenerId := msg.Data.(string)
	delete(this.listener, listenerId)
}

//向通道内写数据
func (this *Event) writeMsg(msg *eventMessage) (ret bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Event.emit err:%v\n", err)
			ret = false
		}
	}()
	select {
	case this.cwrite <- msg:
	default:
		fmt.Printf("event channel full and discard:%v\n", msg)
	}
	return true
}

//Emit 向房间内发送消息
func (this *Event) Emit(data interface{}) bool {
	msg := newEventMessage(eventMessageType_Emit, data)
	return this.writeMsg(msg)
}

//Join 加入房间
func (this *Event) Join(listener eventListener) bool {
	msg := newEventMessage(eventMessageType_Join, listener)
	return this.writeMsg(msg)
}

//Leave 离开房间
func (this *Event) Leave(listenerId string) bool {
	msg := newEventMessage(eventMessageType_Leave, listenerId)
	return this.writeMsg(msg)
}

//Close 关闭房间，如果使用Emitter来管理房间时，请使用Emitter.Close来关闭
func (this *Event) Close() {
	if !atomic.CompareAndSwapInt32(&this.stop, 0, 1) {
		fmt.Printf("Event Stop error:%v\n", this.name)
	}
}
