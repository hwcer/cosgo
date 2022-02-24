package cosmo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strings"
	"time"
)

const (
	MongoTagName     = "bson"
	MongoPrimaryKey  = "_id"
	MongoSetOnInsert = "$setOnInsert"
)

/*
NewClient

uri实例  mongodb://[username:password@]host1[:port1][,host2[:port2],...[,hostN[:portN]]][/[dbname][?options]]

mongodb:// 前缀，代表这是一个Connection String

username:password@ 如果启用了鉴权，需要指定用户密码

hostX:portX 多个 mongos 的地址列表

/dbname 鉴权时，用户帐号所属的数据库

?options 指定额外的连接选项

read preference
    1）primary ： 主节点，默认模式，读操作只在主节点，如果主节点不可用，报错或者抛出异常。

    2）primaryPreferred：首选主节点，大多情况下读操作在主节点，如果主节点不可用，如故障转移，读操作在从节点。

    3）secondary：从节点，读操作只在从节点， 如果从节点不可用，报错或者抛出异常。

    4）secondaryPreferred：首选从节点，大多情况下读操作在从节点，特殊情况（如单主节点架构）读操作在主节点。

    5）nearest：最邻近节点，读操作在最邻近的成员，可能是主节点或者从节点。
*/
func NewClient(address string, opts ...*options.ClientOptions) (client *mongo.Client, err error) {
	if !strings.HasPrefix(address, "mongodb") {
		address = "mongodb://" + address
	}
	c := options.Client().ApplyURI(address)
	c.SetSocketTimeout(time.Second * 5)
	c.SetConnectTimeout(time.Second * 10)
	c.SetServerSelectionTimeout(time.Second * 10)
	client, err = mongo.Connect(context.Background(), append([]*options.ClientOptions{c}, opts...)...)
	if err != nil {
		return
	}
	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		return
	}
	return
}
