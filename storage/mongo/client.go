package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/url"
	"strings"
	"time"
)

const (
	MongoTagName    = "bson"
	MongoPrimarykey = "_id"
)

/*
New
uri实例  mongodb://[username:password@]host1[:port1][,host2[:port2],...[,hostN[:portN]]][/[database][?options]]
mongodb:// 前缀，代表这是一个Connection String
username:password@ 如果启用了鉴权，需要指定用户密码
hostX:portX 多个 mongos 的地址列表
/database 鉴权时，用户帐号所属的数据库
?options 指定额外的连接选项

read preference
    1）primary ： 主节点，默认模式，读操作只在主节点，如果主节点不可用，报错或者抛出异常。

    2）primaryPreferred：首选主节点，大多情况下读操作在主节点，如果主节点不可用，如故障转移，读操作在从节点。

    3）secondary：从节点，读操作只在从节点， 如果从节点不可用，报错或者抛出异常。

    4）secondaryPreferred：首选从节点，大多情况下读操作在从节点，特殊情况（如单主节点架构）读操作在主节点。

    5）nearest：最邻近节点，读操作在最邻近的成员，可能是主节点或者从节点。
*/
func New(address string, opts ...*options.ClientOptions) (*Client, error) {
	if !strings.HasPrefix(address, "mongodb") {
		address = "mongodb://" + address
	}
	var err error
	client := &Client{}
	c := options.Client().ApplyURI(address)
	c.SetSocketTimeout(time.Second * 5)
	c.SetConnectTimeout(time.Second * 10)
	c.SetServerSelectionTimeout(time.Second * 10)
	client.Client, err = mongo.Connect(context.Background(), append([]*options.ClientOptions{c}, opts...)...)
	if err != nil {
		return nil, err
	}
	if err = client.Client.Ping(context.Background(), readpref.Primary()); err != nil {
		return nil, err
	}
	return client, nil
}

//type Collection mongo.Collection

type Client struct {
	*mongo.Client
	url *url.URL
}

func (this *Client) Close() error {
	return this.Client.Disconnect(context.Background())
}

func (this *Client) Database(name string, opts ...*options.DatabaseOptions) *Database {
	return &Database{
		Database: this.Client.Database(name, opts...),
	}
}

func (this *Client) Collection(dbName, collName string, opts ...*options.CollectionOptions) *Collection {
	coll := &Collection{}
	coll.Collection = this.Client.Database(dbName).Collection(collName, opts...)
	return coll
}
