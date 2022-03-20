package updater

import (
	"github.com/hwcer/cosgo/library/logger"
	"sort"
	"time"
)

type wksNew func(handle *Updater) Worker

type WorkerBagType int32

var wksDic = map[WorkerBagType]wksNew{}
var wksNames = map[string]WorkerBagType{}

func Register(bag WorkerBagType, h wksNew, name ...string) {
	wksDic[bag] = h
	if len(name) > 0 && name[0] != "" {
		wksNames[name[0]] = bag
	}
}

type Updater struct {
	uid        string
	wks        map[WorkerBagType]Worker
	time       time.Time
	worker     map[WorkerBagType]bool
	events     map[updaterListenerType][]func(int32)
	overflow   map[int32]int64 //道具溢出,需要使用邮件等其他方式处理
	subVerify  bool
	dataGetter bool
}

func New(uid string) (u *Updater, reset func(uid string), release func()) {
	u = &Updater{
		uid:        uid,
		wks:        make(map[WorkerBagType]Worker),
		worker:     make(map[WorkerBagType]bool),
		events:     make(map[updaterListenerType][]func(int32)),
		overflow:   make(map[int32]int64),
		subVerify:  true,
		dataGetter: false,
	}
	for k, f := range wksDic {
		u.wks[k] = f(u)
	}
	reset = u.reset
	release = u.release
	return
}

//Reset 重置
func (u *Updater) reset(uid string) {
	if u.uid != "" {
		logger.Panic("请不要重复调用Reset")
	}
	u.uid = uid
	u.time = time.Now()
}

func (u *Updater) release() {
	u.uid = ""
	u.overflow = nil
	u.subVerify = true
	u.dataGetter = false
	for k, _ := range u.worker {
		if r, ok := u.wks[k].(release); ok {
			r.release()
		}
	}
	u.worker = make(map[WorkerBagType]bool)
}

func (u *Updater) change(bag int32, mustPullData bool) {
	u.worker[WorkerBagType(bag)] = true
	if mustPullData {
		u.dataGetter = true
	}
}

func (u *Updater) Uid() string {
	return u.uid
}

//Time 获取Updater启动时间
func (u *Updater) Time() time.Time {
	return u.time
}

func (u *Updater) Worker(name string) (w Worker) {
	var bag WorkerBagType
	var ok bool
	if bag, ok = wksNames[name]; !ok {
		return
	}
	w, ok = u.wks[bag]
	return
}

//Subversive true:检查sub, false: 不检查
func (u *Updater) SubVerify(b bool) {
	u.subVerify = b
}

//Keys id为空默认为role.Fields
func (u *Updater) Keys(ids ...int32) {
	for _, id := range ids {
		if w := u.getModuleType(id); w != nil {
			w.Keys(id)
		}
	}
}

func (u *Updater) Val(id int32) (r int64) {
	if w := u.getModuleType(id); w != nil {
		r = w.Val(id)
	}
	return
}

func (u *Updater) Add(id int32, num int32) {
	if num <= 0 {
		logger.Error("Updater.Keys num lte 0: %d", num)
		return
	}
	if w := u.getModuleType(id); w != nil {
		u.Emit(UpdaterListenerTypeAdd, id)
		w.Add(id, int64(num))
		u.dataGetter = false
	}
}

func (u *Updater) Sub(id int32, num int32) {
	if num <= 0 {
		logger.Error("Updater.Sub num lte 0: %d", num)
		return
	}

	if w := u.getModuleType(id); w != nil {
		w.Sub(id, int64(num))
		u.dataGetter = false
	}
}

func (u *Updater) Set(id int32, v interface{}) {
	if w := u.getModuleType(id); w != nil {
		w.Set(id, v)
	}
}

func (u *Updater) Data() (err error) {
	arrModule := u.instance(false)
	for _, w := range arrModule {
		if err = w.Data(); err != nil {
			return
		}
	}
	u.dataGetter = true

	u.Emit(UpdaterListenerTypeFinishData, 0)

	return
}

func (u *Updater) Save() (ret []*Cache, err error) {
	if u.uid == "" {
		return
	}
	if u.dataGetter == false {
		if err = u.Data(); err != nil {
			return
		}
	}

	arrModule := u.instance(true)
	for _, w := range arrModule {
		if err = w.Verify(); err != nil {
			return
		}
	}

	u.Emit(UpdaterListenerTypeBeforeSave, 0)

	for _, w := range arrModule {
		var cache []*Cache
		if cache, err = w.Save(); err != nil {
			return
		} else {
			ret = append(ret, cache...)
		}
	}

	u.Emit(UpdaterListenerTypeFinishSave, 0)

	return
}

func (u *Updater) instance(rank bool) (ret []Worker) {
	for b, _ := range u.worker {
		ret = append(ret, u.wks[b])
	}
	if !rank {
		return
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Sort() < ret[j].Sort()
	})
	return
}

func (u *Updater) getModuleType(id int32) Worker {
	bag := Config.IType(id)
	if bag == 0 {
		logger.Warn("Updater.getModuleType IType not exists: %d", id)
		return nil
	}
	k := WorkerBagType(bag)
	if w, ok := u.wks[k]; ok {
		return w
	} else {
		logger.Warn("Updater.getModuleType worker not exists,id:%v,bag:%v", id, k)
	}
	return nil
}

func (u *Updater) On(typ updaterListenerType, f func(int32)) {
	u.events[typ] = append(u.events[typ], f)
}

func (u *Updater) Emit(typ updaterListenerType, id int32) {
	for _, f := range u.events[typ] {
		f(id)
	}
}
