package confer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/clientv3"
)

var testDir = path.Join(os.Getenv("GOPATH"), "src", "icefire", "confer")

// go test -test.bench=.

func TestViper(t *testing.T) {

	nv := viper.New()
	nv.SetConfigType("yaml") // or viper.SetConfigType("YAML")
	// any approach to require this configuration into your program.
	var yamlExample = []byte(`
Hacker: true
name: steve
hobbies:
- skateboarding
- snowboarding
- go
clothing:
  jacket: leather
  trousers: denim
age: 35
eyes : brown
beard: true
`)

	nv.ReadConfig(bytes.NewBuffer(yamlExample))

	fmt.Println(nv.Get("name"))
}

func TestEtcd(t *testing.T) {
	/*
		endpoints := []string{"http://192.168.66.166:2379"}
		timeOut := time.Second * 8
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: timeOut,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer cli.Close()

		rset := func(v string) {
			_, err = cli.Put(context.TODO(), "/ifconf/zone1002/test.json", v)
			if err != nil {
				log.Fatal(err)
			}
		}

		rset(`{"a":{"b":12345}}`)

		ctx, cancel := context.WithTimeout(context.Background(), timeOut)
		resp, err := cli.Get(ctx, "/ifconf/zone1002/test.json")
		cancel()
		if err != nil {
			log.Fatal(err)
		}
		for _, ev := range resp.Kvs {
			fmt.Printf("%s : %s\n", ev.Key, ev.Value)
		}

		go func() {
			time.Sleep(time.Second * 1)
			rset(`{"a":{"b":12345}}`)
		}()

		rch := cli.Watch(context.Background(), "/ifconf/zone1002/test.json")
		for wresp := range rch {
			for _, ev := range wresp.Events {
				fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}
		}
		time.Sleep(time.Second * 3)
	*/
}

func TestReadLocal(t *testing.T) {
	var val interface{}
	_, err := ReadConf(DEF_HCONF_PROVIDER_LOCAL, ConferParam{
		Name: "example.jsons",
		Path: testDir,
	}, &val)
	assert.Error(t, err)

	vc, err := ReadConf(DEF_HCONF_PROVIDER_LOCAL, ConferParam{
		Name: "example.json",
		Path: testDir,
	}, &val)
	assert.NoError(t, err)

	assert.Equal(t, 12345, vc.GetInt("rate"))
	assert.Equal(t, 5799, vc.GetInt("host.port"))

	var val1 interface{}
	vc1, err := ReadConf(DEF_HCONF_PROVIDER_LOCAL, ConferParam{
		Name: "example.lua",
		Path: testDir,
	}, &val1)
	assert.NoError(t, err)

	assert.Equal(t, 1234, vc1.GetInt("b.c"))
}

func TestReadBuffer(t *testing.T) {
	var val interface{}
	vc, err := ReadConf(DEF_HCONF_PROVIDER_BUFFER, ConferParam{
		Typ:    "json",
		Buffer: []byte(`{"a":{"b":12345}}`),
	}, &val)

	assert.NoError(t, err)
	assert.Equal(t, 12345, vc.GetInt("a.b"))
}

func TestReadRemote(t *testing.T) {
	endpoints := []string{"http://192.168.66.166:2379"}
	timeOut := time.Second * 8
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeOut,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	rset := func(v string) {
		_, err = cli.Put(context.TODO(), "/ifconf/zone1001/test.json", v)
		if err != nil {
			log.Fatal(err)
		}
	}

	rset(`{"a":{"b":1234555555, "c":12333}}`)

	var val interface{}
	vc, err := ReadConf(DEF_HCONF_PROVIDER_REMOTE, ConferParam{
		Name: "/ifconf/zone1001/test.json",
		Addr: "http://192.168.66.166:2379",
	}, &val)

	assert.NoError(t, err)
	assert.Equal(t, 1234555555, vc.GetInt("a.b"))
	assert.NoError(t, vc.Register("a.c", func(ov, nv interface{}) {
		log.Printf("a.c changed:%v,%v", ov, nv)
	}))

	time.Sleep(time.Second * 1)
	rset(`{"a":{"b":1234555555, "c":12333}}`)

	time.Sleep(time.Second * 1)
	rset(`{"a":{"b":123456666, "c":12334}}`)

	time.Sleep(time.Second * 1)
	rset(`{"a":{"b":123456666, "c":[1,2,3,4,5]}}`)

	time.Sleep(time.Second * 2)
	assert.Equal(t, 123456666, vc.GetInt("a.b"))

	time.Sleep(time.Second * 3)

}
