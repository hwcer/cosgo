package main

import (
	"cosgo/grpc/balancer"
	"github.com/gogo/protobuf/proto"
	"log"
	"time"

	"golang.org/x/net/context"
	// 导入grpc包
	"google.golang.org/grpc"
	// 导入刚才我们生成的代码所在的proto包。
	pb "cosgo/grpc/proto"
)

const ETCDName = "project/test"

func main() {
	b := balancer.NewResolver(ETCDName)
	etcd := grpc.RoundRobin(b)
	// 连接grpc服务器
	//conn, err := grpc.Dial("localhost:5005", grpc.WithInsecure())
	conn, err := grpc.Dial(b.Scheme()+"://author/"+svcName, grpc.WithBalancer(etcd), grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// 延迟关闭连接
	defer conn.Close()

	// 初始化Greeter服务客户端
	c := pb.NewGreeterClient(conn)

	// 初始化上下文，设置请求超时时间为1秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 延迟关闭请求会话
	defer cancel()

	// 调用SayHello接口，发送一条消息
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: proto.String("world")})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	// 打印服务的返回的消息
	log.Printf("Greeting: %+v", r)
}
