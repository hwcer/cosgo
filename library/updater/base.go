package updater

type Base struct {
	bag     int32
	sort    int32
	acts    []*Cache
	fields  *fields
	updater *Updater
	verify  bool //是否已经完成verify事件
}

func NewBase(bag, sort int32, updater *Updater) *Base {
	b := &Base{
		bag:     bag,
		sort:    sort,
		acts:    make([]*Cache, 0),
		updater: updater,
		fields:  NewFields(),
	}
	return b
}
func (b *Base) release() {
	b.acts = nil
	b.verify = false
	b.fields.release()
}

func (b *Base) Act(act *Cache, before ...bool) {
	if len(before) > 0 && before[0] {
		b.acts = append([]*Cache{act}, b.acts...)
	} else {
		b.acts = append(b.acts, act)
	}

	b.updater.change(b.bag, false)
}

func (b *Base) Has(key string) bool {
	return b.fields.Has(key)
}

func (b *Base) Bag() int32 {
	return b.bag
}

func (b *Base) Sort() int32 {
	return b.sort
}

func (b *Base) Keys(keys ...interface{}) {
	b.fields.Keys(keys...)
	b.updater.change(b.bag, true)
}

//Fields 使用oid 或者hash key
func (b *Base) Fields(keys ...string) {
	for _, k := range keys {
		b.Keys(k)
	}
}

func (b *Base) Updater() *Updater {
	return b.updater
}
