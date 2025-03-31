package session

import (
	"context"
	"errors"
	"github.com/hwcer/cosgo/redis"
	"github.com/hwcer/cosgo/scc"
	"strings"
	"time"
)

const redisSessionKeyTokenName = "_cookie_key_token"

type Redis struct {
	prefix []string
	client *redis.Client
}

func NewRedis(address interface{}, prefix ...string) (c *Redis, err error) {
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

// Verify 获取session镜像数据
func (this *Redis) Verify(token string) (p *Data, err error) {
	var uuid string
	if uuid, err = Decode(token); err != nil {
		return
	}
	val := map[string]string{}
	rk := this.rkey(uuid)
	if val, err = this.client.HGetAll(context.Background(), rk).Result(); err != nil {
		return
	}
	if v, ok := val[redisSessionKeyTokenName]; !ok {
		return nil, ErrorSessionNotExist
	} else if v != token {
		return nil, ErrorSessionIllegal
	}
	data := map[string]any{}
	for k, v := range val {
		data[k] = v
	}
	//续约
	if Options.MaxAge > 0 {
		this.client.Expire(context.Background(), rk, time.Duration(Options.MaxAge)*time.Second)
	}
	p = NewData(uuid, data, token)
	return
}

func (this *Redis) New(p *Data) error {
	_, err := this.Create(p.uuid, p.Values)
	return err
}

// Create ttl过期时间(s)
func (this *Redis) Create(uuid string, data map[string]any) (p *Data, err error) {
	rk := this.rkey(uuid)
	var id string
	if id, err = Encode(uuid); err != nil {
		return
	}
	var args []any
	for k, v := range data {
		args = append(args, k, v)
	}
	args = append(args, redisSessionKeyTokenName, id)
	if err = this.client.HMSet(context.Background(), rk, args...).Err(); err != nil {
		return
	}
	p = NewData(uuid, data, id)
	return
}

func (this *Redis) Update(p *Data, data map[string]any) (err error) {
	var uuid string
	if uuid, err = Decode(p.id); err != nil {
		return
	}
	rk := this.rkey(uuid)
	//pipeline := this.client.Pipeline()
	if len(data) > 0 {
		args := make([]any, 0, len(data)*2)
		for k, v := range data {
			args = append(args, k, v)
		}
		if _, err = this.client.HMSet(context.Background(), rk, args...).Result(); err != nil {
			return
		}
	}
	return
}

func (this *Redis) Delete(p *Data) (err error) {
	rk := this.rkey(p.uuid)
	_, err = this.client.Del(context.Background(), rk).Result()
	return
}
