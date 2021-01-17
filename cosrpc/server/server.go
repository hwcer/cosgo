package main

import (
	"context"
	"cosgo/cosrpc"
	pb "cosgo/cosrpc/proto"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"log"
	"net"
	"strings"
)

const RPCAddr = "127.0.0.1:5005"
const ETCDAddr = "192.168.66.197:2379"

// 定义server，用来实现proto文件，里面实现的Greeter服务里面的接口
type server struct{}

// 实现SayHello接口
// 第一个参数是上下文参数，所有接口默认都要必填
// 第二个参数是我们定义的HelloRequest消息
// 返回值是我们定义的HelloReply消息，error返回值也是必须的。
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	// 创建一个HelloReply消息，设置Message字段，然后直接返回。
	return &pb.HelloReply{Message: proto.String("Hello " + in.GetName())}, nil
}

func main() {
	// 监听127.0.0.1:50051地址
	lis, err := net.Listen("tcp", RPCAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 实例化grpc服务端
	s := grpc.NewServer()

	// 注册Greeter服务
	pb.RegisterGreeterServer(s, &server{})

	etcd, _ := cosrpc.NewService(strings.Split(ETCDAddr, ";"))
	srv := &cosrpc.Service{
		Name: "test",
		Addr: RPCAddr,
	}
	etcd.Register(srv)
	go etcd.Start()
	defer etcd.Stop()
	// 启动grpc服务
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
