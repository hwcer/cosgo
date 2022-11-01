package bson

const PrimaryKey = "_id"

type Collection struct {
	dict map[string]*Document
}

func (this *Collection) Has(id string) (r bool) {
	_, r = this.dict[id]
	return
}
func (this *Collection) Get(id string) (r *Document) {
	r, _ = this.dict[id]
	return
}

func (this *Collection) Remove(id ...string) {
	for _, k := range id {
		delete(this.dict, k)
	}
}
func (this *Collection) Range(f func(string, *Document) bool) {
	for k, v := range this.dict {
		if !f(k, v) {
			return
		}
	}
}

// Insert 插入对象
// values  bson二进制或者可以bson.Marshal的 struct、map
func (this *Collection) Insert(values ...interface{}) error {
	r := make(map[string]*Document, len(values))
	for _, i := range values {
		if v, err := Marshal(i); err != nil {
			return err
		} else {
			ele := v.Element(PrimaryKey)
			if ele == nil {
				return ErrorNoPrimaryKey
			} else {
				r[ele.GetString()] = v
			}
		}
	}
	for k, v := range r {
		this.dict[k] = v
	}
	return nil
}

// Unmarshal 使用i解析id对应的文档  id可以是  100.name 格式 来解析id=100的name字段
func (this *Collection) Unmarshal(id string, i interface{}) (err error) {
	k1, k2 := Split(id)
	if v := this.Get(k1); v != nil {
		err = v.Unmarshal(k2, i)
	}
	return
}
