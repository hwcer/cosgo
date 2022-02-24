package updater

import "strconv"

var Config = struct {
	IMax    func(iid int32) (r int64)
	IType   func(iid int32) (bag int32)
	Upgrade Upgrade //lv->exp

}{
	IMax:  itemMax,
	IType: itemType,
}

type Upgrade interface {
	Exp(lv string) (key string) //根据lv获取exp
	Upgrade(lv, exp int64) (newlv, newexp int64)
}

//itemMax 物品最大拥有数量,仅限于hmap类型
func itemMax(iid int32) (r int64) {
	return
}

//itemType 通过iid获取bag
func itemType(iid int32) (bag int32) {
	if iid < 10 {
		return
	}
	s := strconv.Itoa(int(iid))
	if b, err := strconv.Atoi(s[0:2]); err != nil {
		return
	} else {
		bag = int32(b)
	}
	return
}
