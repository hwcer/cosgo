// Package session 提供会话管理功能，支持内存和Redis存储
package session

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/hwcer/cosgo/redis"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/logger"
)

type Redis struct {
	prefix []string
	client *redis.Client
}

func NewRedis(address any, prefix ...string) (c *Redis, err error) {
	// 精确分配容量,避免rkey中append时复用并并发写入底层数组
	c = &Redis{
		prefix: make([]string, 0, len(prefix)+1),
	}
	c.prefix = append(c.prefix, prefix...)
	c.prefix = append(c.prefix, "cookie")

	switch v := address.(type) {
	case *redis.Client:
		c.client = v
	case string:
		err = c.init(v)
	default:
		err = errors.New("address type must be string or *redis.Client")
	}
	return
}

func (this *Redis) init(address string) (err error) {
	if this.client != nil {
		return
	}
	this.client, err = redis.New(address)
	if err != nil {
		return err
	}
	scc.Trigger(func() {
		_ = this.client.Close()
	})
	return
}

func (this *Redis) rkey(uuid string) string {
	return strings.Join(append(this.prefix, uuid), "-")
}

// Get 获取session镜像数据
func (this *Redis) Get(uuid string) (p *Data, err error) {
	rk := this.rkey(uuid)
	var val map[string]string
	if val, err = this.client.HGetAll(context.Background(), rk).Result(); err != nil {
		return
	}
	data := map[string]any{}
	for k, v := range val {
		data[k] = v
	}
	//续约
	if Options.MaxAge > 0 {
		if e := this.client.Expire(context.Background(), rk, time.Duration(Options.MaxAge)*time.Second).Err(); e != nil {
			logger.Alert("session.redis Expire renew failed uuid=%s: %v", uuid, e)
		}
	}
	p = NewData(uuid, data)
	return
}

func (this *Redis) New(p *Data) error {
	_, err := this.Create(p.uuid, p.Values())
	return err
}

// Create ttl过期时间(s)
func (this *Redis) Create(uuid string, data map[string]any) (p *Data, err error) {
	rk := this.rkey(uuid)
	//🔴 新会话先清旧键:HMSET 是合并语义,同 uuid 二次登录(新设备/被顶后重登)时
	//存储里残留的旧 uid、旧 selector 等字段会被新会话的 Verify 原样还原——
	//新设备未选角即继承旧角色的身份状态
	if err = this.client.Del(context.Background(), rk).Err(); err != nil {
		return
	}
	var args []any
	for k, v := range data {
		args = append(args, k, v)
	}
	if len(args) == 0 {
		//🔴 空会话占位:HMSET 不接受零字段(Redis 8 报 wrong number of arguments),
		//但"无任何 cookies 的登录"是合法场景(gateway players.Create 的空 values)
		args = append(args, "_e", 1)
	}
	if err = this.client.HMSet(context.Background(), rk, args...).Err(); err != nil {
		return
	}
	// 设置过期时间
	if Options.MaxAge > 0 {
		if e := this.client.Expire(context.Background(), rk, time.Duration(Options.MaxAge)*time.Second).Err(); e != nil {
			logger.Alert("session.redis Expire set failed uuid=%s: %v", uuid, e)
		}
	}
	p = NewData(uuid, data)
	return
}

func (this *Redis) Update(p *Data, data map[string]any) (err error) {
	uuid := p.UUID()
	rk := this.rkey(uuid)
	//pipeline := this.client.Pipeline()
	if len(data) > 0 {
		args := make([]any, 0, len(data)*2)
		for k, v := range data {
			args = append(args, k, v)
		}
		_, err = this.client.HMSet(context.Background(), rk, args...).Result()
	}
	// 更新过期时间
	if Options.MaxAge > 0 {
		if e := this.client.Expire(context.Background(), rk, time.Duration(Options.MaxAge)*time.Second).Err(); e != nil {
			logger.Alert("session.redis Expire update failed uuid=%s: %v", uuid, e)
		}
	}
	return
}

func (this *Redis) Delete(p *Data) (err error) {
	rk := this.rkey(p.uuid)
	_, err = this.client.Del(context.Background(), rk).Result()
	return
}

// DeleteKeys 写穿键删除(HDEL),实现 StorageDeleter 可选接口。
// 删除不做 Expire 续约:删除语义不应延长会话寿命
func (this *Redis) DeleteKeys(p *Data, keys ...string) (err error) {
	if len(keys) == 0 {
		return nil
	}
	rk := this.rkey(p.uuid)
	_, err = this.client.HDel(context.Background(), rk, keys...).Result()
	return
}
