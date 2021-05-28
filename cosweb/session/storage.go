package session

import (
	"errors"
	"github.com/hwcer/cosgo/utils"
	"strconv"
	"sync"
	"sync/atomic"
)

const storageKeyBitSize = 36

var StorageExpireKey = "_StorageExpireKey"

func NewStorage(data map[string]interface{}) *Storage {
	return &Storage{userdata: data}
}

type Storage struct {
	key      string
	mutex    sync.Mutex
	locked   int32
	expire   int64
	userdata map[string]interface{}
}

func (this *Storage) Set(val map[string]interface{}) {
	this.mutex.Lock()
	for k, v := range val {
		this.userdata[k] = v
	}
	this.mutex.Unlock()
}
func (this *Storage) Get() map[string]interface{} {
	val := make(map[string]interface{}, len(this.userdata))
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for k, v := range this.userdata {
		val[k] = v
	}
	return val
}
func (this *Storage) Lock() bool {
	return atomic.CompareAndSwapInt32(&this.locked, 0, 1)
}
func (this *Storage) UnLock() {
	atomic.StoreInt32(&this.locked, 0)
}

//session id
func (this *Storage) GetStorageKey() string {
	return this.key
}

func (this *Storage) GetArrayMapKey() utils.ArrayMapKey {
	arrayMapKey, _ := ArrayMapKeyDecode(this.key)
	return arrayMapKey
}

func (this *Storage) SetArrayMapKey(arrayMapKey utils.ArrayMapKey) error {
	if this.key != "" {
		return errors.New("session Storage key exist")
	}
	this.key = ArrayMapKeyEncode(arrayMapKey)
	return nil
}

func ArrayMapKeyEncode(arrayMapKey utils.ArrayMapKey) string {
	return strconv.FormatInt(int64(arrayMapKey), storageKeyBitSize)
}

func ArrayMapKeyDecode(key string) (utils.ArrayMapKey, error) {
	num, err := strconv.ParseInt(key, 10, storageKeyBitSize)
	if err != nil {
		return 0, err
	} else {
		return utils.ArrayMapKey(num), nil
	}
}
