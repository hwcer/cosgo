package redis

import (
	"github.com/go-redis/redis"
	"github.com/hwcer/cosgo/session/options"
	"github.com/hwcer/cosgo/utils"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const redisSessionKeyUid = "_s_uid"
const redisSessionKeyLock = "_s_lock"

var Options = &options.Options //导入配置

type Redis struct {
	prefix  string
	client  *redis.Client
	address *url.URL
}

func New(address string) (*Redis, error) {
	addr, err := utils.NewUrl(address, "tcp")
	if err != nil {
		return nil, err
	}
	c := &Redis{
		prefix:  "cookie",
		address: addr,
	}
	return c, nil
}

func (this *Redis) Start() (err error) {
	//rdb := redis.NewFailoverClient(&redis.FailoverOptions{
	//	MasterName:    "master",
	//	SentinelAddrs: []string{"x.x.x.x:26379", "xx.xx.xx.xx:26379", "xxx.xxx.xxx.xxx:26379"},
	//})
	//
	//rdb := redis.NewClusterClient(&redis.ClusterOptions{
	//	Addrs: []string{":7000", ":7001", ":7002", ":7003", ":7004", ":7005"},
	//})

	opts := &redis.Options{
		Addr:    this.address.Host,
		Network: this.address.Scheme,
	}
	query := this.address.Query()
	opts.Password = query.Get("password")
	if db := query.Get("db"); db != "" {
		if opts.DB, err = strconv.Atoi(db); err != nil {
			return err
		}
	}

	this.client = redis.NewClient(opts)
	_, err = this.client.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func (this *Redis) Close() error {
	this.client.Close()
	return nil
}

func (this *Redis) rkey(uid string) string {
	return strings.Join([]string{this.prefix, uid}, "-")
}

func (this *Redis) lock(rkey string, data map[string]string) bool {
	var num int64
	var err error
	if num, err = strconv.ParseInt(data[redisSessionKeyLock], 10, 10); err != nil || num > 0 {
		return false
	}
	if num, err = this.client.HIncrBy(rkey, redisSessionKeyLock, 1).Result(); err != nil || num > 1 {
		return false
	}
	return true
}

//获取session镜像数据
func (this *Redis) Get(key string, lock bool) (uid string, data map[string]interface{}, err error) {
	var ok bool
	var val map[string]string
	rkey := this.rkey(key)
	if val, err = this.client.HGetAll(rkey).Result(); err != nil {
		return
	}
	if uid, ok = val[redisSessionKeyUid]; !ok {
		return "", nil, options.ErrorSessionNotExist
	}
	if lock && !this.lock(rkey, val) {
		return "", nil, options.ErrorSessionLocked
	}
	data = make(map[string]interface{}, len(val))
	for k, v := range val {
		data[k] = v
	}
	return
}

func (this *Redis) Create(uid string, data map[string]interface{}, expire int64, lock bool) (sid, key string, err error) {
	key = uid
	data[redisSessionKeyUid] = uid
	if lock {
		data[redisSessionKeyLock] = 1
	}
	rkey := this.rkey(uid)
	//pipeline := this.client.Pipeline()
	if err = this.client.HMSet(rkey, data).Err(); err != nil {
		return
	}

	if expire > 0 {
		if err = this.client.ExpireAt(rkey, time.Unix(expire, 0)).Err(); err != nil {
			return
		}
	}
	//if err = pipeline.FlushAll().Err(); err != nil {
	//	return
	//}
	//if err = pipeline.Save().Err(); err != nil {
	//	return
	//}
	sid, err = options.Encode(uid)
	return

}
func (this *Redis) Save(key string, data map[string]interface{}, expire int64, unlock bool) (err error) {
	rkey := this.rkey(key)
	if data == nil {
		data = make(map[string]interface{})
	}
	//pipeline := this.client.Pipeline()
	if unlock {
		data[redisSessionKeyLock] = int64(0)
	}

	if len(data) > 0 {
		if _, err = this.client.HMSet(rkey, data).Result(); err != nil {
			return
		}
	}
	if expire > 0 {
		if _, err = this.client.ExpireAt(rkey, time.Unix(expire, 0)).Result(); err != nil {
			return
		}
	}

	//_, err = pipeline.Save().Result()

	return
}

func (this *Redis) Delete(key string) (err error) {
	rkey := this.rkey(key)
	_, err = this.client.Del(rkey).Result()
	return
}
