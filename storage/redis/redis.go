package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/hwcer/cosgo/utils"
	"net/url"
	"strconv"
)

type Client struct {
	*redis.Client
	address string
}

func IsNil(err error) bool {
	return err.Error() == "redis: nil"
}

func (c *Client) Select(db int) (client *Client, err error) {
	address := c.Address(db)
	return New(address)
}

func (c *Client) Address(db int) string {
	uri, _ := url.Parse(c.address)
	query := uri.Query()
	query.Set("db", strconv.Itoa(db))
	uri.RawQuery = query.Encode()
	return uri.String()
}

func New(address string) (client *Client, err error) {
	var uri *url.URL
	uri, err = utils.NewUrl(address, "tcp")
	if err != nil {
		return
	}
	opts := &redis.Options{
		Addr:    uri.Host,
		Network: uri.Scheme,
	}
	query := uri.Query()
	opts.Password = query.Get("password")
	if db := query.Get("db"); db != "" {
		if opts.DB, err = strconv.Atoi(db); err != nil {
			return
		}
	}
	c := redis.NewClient(opts)

	_, err = c.Ping(context.Background()).Result()
	if err != nil {
		return
	}
	client = &Client{Client: c, address: uri.String()}
	return
}
