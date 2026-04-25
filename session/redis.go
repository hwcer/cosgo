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
	c = &Redis{
		prefix: prefix,
	}
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
	val := map[string]string{}
	rk := this.rkey(uuid)
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
	var args []any
	for k, v := range data {
		args = append(args, k, v)
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
