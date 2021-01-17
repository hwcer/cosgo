package cosrpc

import (
	"context"
	"encoding/json"
	"errors"
	"go.etcd.io/etcd/clientv3"
	"log"
	"time"
)

var Etcd *Registry

// Etcd服务注册
type Service struct {
	Name string
	Addr string
}

func (this *Service) Key() string {
	return this.Name + "/" + this.Addr
}
func (this *Service) Value() string {
	val, _ := json.Marshal(this)
	return string(val)
}

type Registry struct {
	stop    chan error
	lease   *clientv3.LeaseGrantResponse //租约
	Client  *clientv3.Client             //客户端
	service []*Service
}

// NewService 创建一个注册服务
func NewService(endpoints []string) (registry *Registry, err error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 200,
	})

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	registry = &Registry{
		Client: client,
	}
	return
}

// Start 注册服务启动
func (this *Registry) Start() (err error) {
	ch, err := this.keepAlive()
	if err != nil {
		log.Fatal(err)
		return
	}

	for {
		select {
		case err := <-this.stop:
			return err
		case <-this.Client.Ctx().Done():
			return errors.New("service closed")
		case resp, ok := <-ch:
			// 监听租约
			if !ok {
				log.Println("keep alive channel closed")
				return this.revoke()
			}
			log.Printf("Recv reply from service, ttl:%d", resp.TTL)
		}
	}

	return
}

func (this *Registry) Stop() {
	this.stop <- nil
}

func (this *Registry) Register(s *Service) {
	this.service = append(this.service, s)
}
func (this *Registry) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	// 创建一个租约
	resp, err := this.Client.Grant(context.TODO(), 5)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	for _, s := range this.service {
		_, err = this.Client.Put(context.TODO(), s.Key(), s.Value(), clientv3.WithLease(resp.ID))
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	this.lease = resp
	return this.Client.KeepAlive(context.TODO(), resp.ID)
}

func (this *Registry) revoke() error {
	_, err := this.Client.Revoke(context.TODO(), this.lease.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("servide:%s stop\n")
	return err
}
