package updater

import (
	"github.com/hwcer/cosgo/library/logger"
)

//Table 不可以叠加的物品,卡牌，装备...
type Table struct {
	*HMap
	unique bool //道具是否唯一，如果唯一需要在model中提供分解方案
}

func NewTable(bag, sort int32, model modelTable, updater *Updater, unique bool) *Table {
	if unique {
		if _, ok := model.(resolve); !ok {
			logger.Panic("NewTable设置为唯一但没有提供分解方案")
		}
	}
	i := &Table{
		unique: unique,
	}
	i.HMap = NewHMap(bag, sort, model, updater)
	i.release()
	return i
}

//Sub 等价于删除
func (this *Table) Sub(k int32, v int64) {
	logger.Error("bulkWrite.table无法使用Sub(k int32, v int64),请使用Del(string)")
}

//Set 设置道具属性，唯一道具可以使用iid或者oid操作，，非唯一道具只能使用OID操作
func (this *Table) Set(id interface{}, v interface{}) {
	var ok bool
	var oid string
	var val map[string]interface{}
	if val, ok = v.(map[string]interface{}); !ok {
		logger.Error("table set v error")
		return
	}

	if _, ok = id.(string); !ok && !this.unique {
		logger.Error("updater table set 道具不唯一,只能使用oid操作物品:%v", id)
		return
	}

	iid, oid, err := this.HMap.parseId(id)
	if err != nil {
		logger.Error("%v", err)
		return
	}

	act := &Cache{ID: oid, Id: iid, T: ActTypeSet, K: "*", V: val, B: this.Base.Bag()}
	this.Act(act)
	this.Keys(oid)
}

func (this *Table) Add(iid int32, val int64) {
	if this.unique {
		act := &Cache{ID: "", Id: iid, T: ActTypeAdd, K: "", V: val, B: this.Base.Bag()}
		this.Act(act)
		this.Keys(iid)
		if rid, _, ok := this.model.(resolve).Resolve(iid, 1); ok {
			this.Keys(rid)
		}
	} else {
		for i := int64(0); i < val; i++ {
			act := &Cache{ID: "", Id: iid, T: ActTypeNew, K: "", V: 1, B: this.Base.Bag()}
			this.Act(act)
		}
	}

}
func (this *Table) Del(oid string) {
	act := &Cache{ID: oid, Id: 0, T: ActTypeDel, K: "", V: 0, B: this.Base.Bag()}
	this.Act(act)
}

func (this *Table) Act(act *Cache) {
	this.Base.Act(act)
	if act.T == ActTypeSet {
		this.Keys(act.ID)
	}
	if this.Base.verify {
		this.Parse(act)
	}
}

func (this *Table) Verify() (err error) {
	this.Base.verify = true
	if len(this.Base.acts) == 0 {
		return nil
	}
	for _, act := range this.Base.acts {
		if err = this.Parse(act); err != nil {
			return
		}
	}
	return nil
}
